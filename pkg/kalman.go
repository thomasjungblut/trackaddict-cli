// this is from https://github.com/slobdell/kalman-filter-example
// which sadly can't be imported as a dependency since it only has a main package

package pkg

import (
	"github.com/slobdell/basicMatrix"
)

/*
	Although these variables aren't expressive, they're based on existing mathematical conventions
	and in reality should be completely abstract.  The variables in my own words expressed below:

	H: For our usage, this should just be an identity matrix.  In practice this is meant to be
	a transformation matrix to standardize inputs to the system, but I'm enforcing this in the
	API itself; This should simplify usage and a bit of performance by not having to use this

	P: Newest estimate for average error for each part of state. This value will evolve internally
	from the kalman filter, so initializing as an identity matrix is also acceptable

	Q: Abstractly, the process error variance.  Explicitly for our use case, this is the covariance
	matrix for the accelerometer.  To find, you can leave the accelerometer at rest and take the standard
	deviation, then square that for the variance.  Matrix would then be

	[AVariance 0]
	[0 AVariance]

	Additionally, when computing standard deviation, in this context it would make sense to override
	the mean value of the readings to be 0 to account for a blatant offset from the sensor.

	R: Abstractly, the measurement error variance. Explicitly for our use case, this is the covariance
	matrix of the GPS.  If you can get the actual standard deviation of the GPS, this might work, but
	if you take GPS readings at rest, you might have a GPS lock that results in extremely minimal error.

	In practice, I just took the advertised +/- value from the GPS (i.e. uBlock is accurate +/- 1 meter allegedly, so you can use that).

	u: Overridden during each prediction step; Setting as a struct attribute for performance reasons. This
	is the input matrix of high frequency sensor readings that without subject to any error would give us
	an accurate state of the world.

	In our case, it's a 1x1 matrix of accelerometer input in a given direction.

	z: Overridden during each prediction step; Setting as a struct attribute for performance reasongs. This
	is the input matrix of low frequency sensor readings that are absolute but presumably high standard
	deviation.

	In our case, it's a 2x1 matrix of GPS position and velocity
	[ P
	  v ]

	  A: The state transition matrix. Abstractly, this is a matrix that defines a set of of equations that define what the next step would like given no additional inputs but a "next step" (or more than likely, change in time). Given that this struct is explicitly for fusing position and acceleration, it's:

	  [ 1 t
	    0 1 ]

	To explain the above, if you have position, then its next position is the previous position + current velocity * times. If you have velocity, then its next velocity will be the current velocity.

	B: Control matrix. Given input changes to the system, this matrix multiplied by the input will present new deltas
	to the current state.  In our case, these are the equations needed to handle input acceleration.  Specifically:

	[ 0.5t^2
	  t     ]
*/
type KalmanFilterFusedPositionAccelerometer struct {
	I                            *basicMatrix.Matrix // identity matrix used in some calculations
	H                            *basicMatrix.Matrix // transformation matrix for input data
	P                            *basicMatrix.Matrix // initial guess for covariance
	Q                            *basicMatrix.Matrix // process (accelerometer) error variance
	R                            *basicMatrix.Matrix // measurement (GPS) error variance
	u                            *basicMatrix.Matrix // INPUT control (accelerometer) matrix
	z                            *basicMatrix.Matrix // INPUT measurement (GPS) matrix
	A                            *basicMatrix.Matrix // State Transition matrix
	B                            *basicMatrix.Matrix // Control matrix
	currentState                 *basicMatrix.Matrix
	currentStateTimestampSeconds float64
}

func (k *KalmanFilterFusedPositionAccelerometer) Predict(accelerationThisAxis, timestampNow float64) {
	deltaT := timestampNow - k.currentStateTimestampSeconds

	k.recreateControlMatrix(deltaT)
	k.recreateStateTransitionMatrix(deltaT)

	k.u.Put(0, 0, accelerationThisAxis)

	k.currentState = (k.A.MultipliedBy(k.currentState)).Add(k.B.MultipliedBy(k.u))

	k.P = ((k.A.MultipliedBy(k.P)).MultipliedBy(k.A.Transpose())).Add(k.Q)

	k.currentStateTimestampSeconds = timestampNow
}

func (k *KalmanFilterFusedPositionAccelerometer) Update(position float64, velocityThisAxis float64, positionError *float64, velocityError float64) {

	k.z.Put(0, 0, float64(position))
	k.z.Put(1, 0, float64(velocityThisAxis))

	if positionError != nil {
		k.R.Put(0, 0, *positionError**positionError)
	} else {
		k.R.Put(1, 1, velocityError*velocityError)
	}

	y := k.z.Subtract(k.currentState)
	s := k.P.Add(k.R)
	sInverse, err := s.Inverse()
	if err != nil {
		// matrix has no inverse, abort
		return
	}
	K := k.P.MultipliedBy(sInverse)

	k.currentState = k.currentState.Add(K.MultipliedBy(y))

	k.P = (k.I.Subtract(K)).MultipliedBy(k.P)

	/*
		above is equivalent to:
			updatedP := k.P.Subtract(K.MultipliedBy(k.P))
		which would explain some confusion on the internets
	*/
}

func (k *KalmanFilterFusedPositionAccelerometer) recreateControlMatrix(deltaSeconds float64) {
	dtSquared := 0.5 * deltaSeconds * deltaSeconds

	k.B.Put(0, 0, dtSquared)
	k.B.Put(1, 0, deltaSeconds)
}
func (k *KalmanFilterFusedPositionAccelerometer) recreateStateTransitionMatrix(deltaSeconds float64) {
	k.A.Put(0, 0, 1.0)
	k.A.Put(0, 1, deltaSeconds)

	k.A.Put(1, 0, 0.0)
	k.A.Put(1, 1, 1.0)
}

func (k *KalmanFilterFusedPositionAccelerometer) GetPredictedPosition() float64 {
	return k.currentState.Get(0, 0)
}

func (k *KalmanFilterFusedPositionAccelerometer) GetPredictedVelocityThisAxis() float64 {
	return k.currentState.Get(1, 0)
}

func NewKalmanFilterFusedPositionAccelerometer(initialPosition float64,
	positionStandardDeviation float64,
	accelerometerStandardDeviation float64,
	currentTimestampSeconds float64) *KalmanFilterFusedPositionAccelerometer {

	currentState := basicMatrix.NewMatrix(2, 1)

	currentState.Put(0, 0, float64(initialPosition))
	currentState.Put(1, 0, 0.0)

	u := basicMatrix.NewMatrix(1, 1)
	z := basicMatrix.NewMatrix(2, 1)
	H := basicMatrix.NewIdentityMatrix(2, 2)
	P := basicMatrix.NewIdentityMatrix(2, 2)
	I := basicMatrix.NewIdentityMatrix(2, 2)

	Q := basicMatrix.NewMatrix(2, 2)
	Q.Put(0, 0, accelerometerStandardDeviation*accelerometerStandardDeviation)
	Q.Put(1, 1, accelerometerStandardDeviation*accelerometerStandardDeviation)

	R := basicMatrix.NewMatrix(2, 2)
	R.Put(0, 0, float64(positionStandardDeviation*positionStandardDeviation))
	R.Put(1, 1, float64(positionStandardDeviation*positionStandardDeviation))
	B := basicMatrix.NewMatrix(2, 1)
	A := basicMatrix.NewMatrix(2, 2)

	return &KalmanFilterFusedPositionAccelerometer{
		I:                            I,
		A:                            A,
		B:                            B,
		z:                            z,
		u:                            u,
		H:                            H,
		P:                            P,
		Q:                            Q,
		R:                            R,
		currentState:                 currentState,
		currentStateTimestampSeconds: currentTimestampSeconds,
	}
}
