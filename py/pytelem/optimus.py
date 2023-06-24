# hyperspeed forward and backwards analytics engine


import numpy as np

from numba import jit


# import jax.numpy as np


# TODO: define 3d vector space - x,y,z oriented around car/world?
# solvers should not care about position but we should be able to convert
# transforms between car/world spaces.

# let's define x as the forward-backward axis (with forward being positive)
# y as the lateral axis, with right being positive
# and z being the vertical, with up being positive.

# for simplicities’ sake, the car only rotates on the y-axis (up and down hills).
# z rotation can be determined from distance along the route.

# some data (wind) is only a 2d vector at a given time point.


def fsolve_discrete():
    ...


def dist_to_pos(dist: float):
    "convert a distance along the race path to a position in 3d space"


### All units are BASE SI (no prefix except for kilogram)
ATM_MOLAR_MASS = 0.0289644  # kg/mol
STANDARD_TEMP = 288.15  # K
STANDARD_PRES = 101325.0  # Pa

AIR_GAS_CONSTANT = 8.31432  # N*m/s^2

EARTH_TEMP_LAPSE = -0.0065
EARTH_GRAVITY = 9.80665  # m/s^2
EARTH_RADIUS = 6378140.0  # m
EARTH_AXIS_INCLINATION = 23.45  # degrees


# FIXME: use named constants here


@jit
def get_pressure_el(
        el: float,
        Ps=STANDARD_PRES,
        Ts: float = STANDARD_TEMP,
        T_lapse: float = EARTH_TEMP_LAPSE,
):
    """Gets the pressure at a point given eleveation - assumes
    standard pressure,temperature, gas constants, etc"""

    return Ps * (Ts / (Ts + T_lapse * el)) ** (
            (ATM_MOLAR_MASS * EARTH_GRAVITY) / (AIR_GAS_CONSTANT / T_lapse)
    )


@jit
def estimate_temp(el: float, Ts: float = STANDARD_TEMP, T_lapse=EARTH_TEMP_LAPSE):
    return Ts + el * T_lapse


def make_cubic(a, b, c, d):
    """returns a simple cubic function"""

    def poly(x):
        return a + b * x + c * (x ** 2) + (x ** 3) / d

    return jit(poly)


@jit
@vmap
def get_radiation_direct(yday, altitude_deg):
    """Calculate the direct radiation at a given day of the year given the angle of the sun
    from the horizon."""

    flux = 1160 + (75 * np.sin(2 * np.pi / 365 * (yday - 275)))

    optical_depth = 0.174 + (0.035 * np.sin(2 * np.pi / 365 * (yday - 100)))

    air_mass_ratio = 1 / np.sin(np.radians(altitude_deg))
    # from Masters, p. 412

    return flux * np.exp(-1 * optical_depth * air_mass_ratio) * (altitude_deg > 0)


# We start by defining MANY constants.
# to skip this, Ctrl-F to END COEFF
# START COEFF


# heliocentric longitude, latitude, radius (section 3.2) coefficients

# heliocentric longitude coefficients
L0_TABLE = np.array(
    [
        [175347046.0, 0.0, 0.0],
        [3341656.0, 4.6692568, 6283.07585],
        [34894.0, 4.6261, 12566.1517],
        [3497.0, 2.7441, 5753.3849],
        [3418.0, 2.8289, 3.5231],
        [3136.0, 3.6277, 77713.7715],
        [2676.0, 4.4181, 7860.4194],
        [2343.0, 6.1352, 3930.2097],
        [1324.0, 0.7425, 11506.7698],
        [1273.0, 2.0371, 529.691],
        [1199.0, 1.1096, 1577.3435],
        [990.0, 5.233, 5884.927],
        [902.0, 2.045, 26.298],
        [857.0, 3.508, 398.149],
        [780.0, 1.179, 5223.694],
        [753.0, 2.533, 5507.553],
        [505.0, 4.583, 18849.228],
        [492.0, 4.205, 775.523],
        [357.0, 2.92, 0.067],
        [317.0, 5.849, 11790.629],
        [284.0, 1.899, 796.298],
        [271.0, 0.315, 10977.079],
        [243.0, 0.345, 5486.778],
        [206.0, 4.806, 2544.314],
        [205.0, 1.869, 5573.143],
        [202.0, 2.458, 6069.777],
        [156.0, 0.833, 213.299],
        [132.0, 3.411, 2942.463],
        [126.0, 1.083, 20.775],
        [115.0, 0.645, 0.98],
        [103.0, 0.636, 4694.003],
        [102.0, 0.976, 15720.839],
        [102.0, 4.267, 7.114],
        [99.0, 6.21, 2146.17],
        [98.0, 0.68, 155.42],
        [86.0, 5.98, 161000.69],
        [85.0, 1.3, 6275.96],
        [85.0, 3.67, 71430.7],
        [80.0, 1.81, 17260.15],
        [79.0, 3.04, 12036.46],
        [75.0, 1.76, 5088.63],
        [74.0, 3.5, 3154.69],
        [74.0, 4.68, 801.82],
        [70.0, 0.83, 9437.76],
        [62.0, 3.98, 8827.39],
        [61.0, 1.82, 7084.9],
        [57.0, 2.78, 6286.6],
        [56.0, 4.39, 14143.5],
        [56.0, 3.47, 6279.55],
        [52.0, 0.19, 12139.55],
        [52.0, 1.33, 1748.02],
        [51.0, 0.28, 5856.48],
        [49.0, 0.49, 1194.45],
        [41.0, 5.37, 8429.24],
        [41.0, 2.4, 19651.05],
        [39.0, 6.17, 10447.39],
        [37.0, 6.04, 10213.29],
        [37.0, 2.57, 1059.38],
        [36.0, 1.71, 2352.87],
        [36.0, 1.78, 6812.77],
        [33.0, 0.59, 17789.85],
        [30.0, 0.44, 83996.85],
        [30.0, 2.74, 1349.87],
        [25.0, 3.16, 4690.48],
    ]
)
L1_TABLE = np.array(
    [
        [628331966747.0, 0.0, 0.0],
        [206059.0, 2.678235, 6283.07585],
        [4303.0, 2.6351, 12566.1517],
        [425.0, 1.59, 3.523],
        [119.0, 5.796, 26.298],
        [109.0, 2.966, 1577.344],
        [93.0, 2.59, 18849.23],
        [72.0, 1.14, 529.69],
        [68.0, 1.87, 398.15],
        [67.0, 4.41, 5507.55],
        [59.0, 2.89, 5223.69],
        [56.0, 2.17, 155.42],
        [45.0, 0.4, 796.3],
        [36.0, 0.47, 775.52],
        [29.0, 2.65, 7.11],
        [21.0, 5.34, 0.98],
        [19.0, 1.85, 5486.78],
        [19.0, 4.97, 213.3],
        [17.0, 2.99, 6275.96],
        [16.0, 0.03, 2544.31],
        [16.0, 1.43, 2146.17],
        [15.0, 1.21, 10977.08],
        [12.0, 2.83, 1748.02],
        [12.0, 3.26, 5088.63],
        [12.0, 5.27, 1194.45],
        [12.0, 2.08, 4694.0],
        [11.0, 0.77, 553.57],
        [10.0, 1.3, 6286.6],
        [10.0, 4.24, 1349.87],
        [9.0, 2.7, 242.73],
        [9.0, 5.64, 951.72],
        [8.0, 5.3, 2352.87],
        [6.0, 2.65, 9437.76],
        [6.0, 4.67, 4690.48],
    ]
)
L2_TABLE = np.array(
    [
        [52919.0, 0.0, 0.0],
        [8720.0, 1.0721, 6283.0758],
        [309.0, 0.867, 12566.152],
        [27.0, 0.05, 3.52],
        [16.0, 5.19, 26.3],
        [16.0, 3.68, 155.42],
        [10.0, 0.76, 18849.23],
        [9.0, 2.06, 77713.77],
        [7.0, 0.83, 775.52],
        [5.0, 4.66, 1577.34],
        [4.0, 1.03, 7.11],
        [4.0, 3.44, 5573.14],
        [3.0, 5.14, 796.3],
        [3.0, 6.05, 5507.55],
        [3.0, 1.19, 242.73],
        [3.0, 6.12, 529.69],
        [3.0, 0.31, 398.15],
        [3.0, 2.28, 553.57],
        [2.0, 4.38, 5223.69],
        [2.0, 3.75, 0.98],
    ]
)
L3_TABLE = np.array(
    [
        [289.0, 5.844, 6283.076],
        [35.0, 0.0, 0.0],
        [17.0, 5.49, 12566.15],
        [3.0, 5.2, 155.42],
        [1.0, 4.72, 3.52],
        [1.0, 5.3, 18849.23],
        [1.0, 5.97, 242.73],
    ]
)
L4_TABLE = np.array([[114.0, 3.142, 0.0], [8.0, 4.13, 6283.08], [1.0, 3.84, 12566.15]])
L5_TABLE = np.array([[1.0, 3.14, 0.0]])

HELIO_L = [L0_TABLE, L1_TABLE, L2_TABLE, L3_TABLE, L4_TABLE, L5_TABLE]

# heliocentric latitude coefficients
B0_TABLE = np.array(
    [
        [280.0, 3.199, 84334.662],
        [102.0, 5.422, 5507.553],
        [80.0, 3.88, 5223.69],
        [44.0, 3.7, 2352.87],
        [32.0, 4.0, 1577.34],
    ]
)
B1_TABLE = np.array([[9.0, 3.9, 5507.55], [6.0, 1.73, 5223.69]])

HELIO_B = [B0_TABLE, B1_TABLE]

# heliocentric radius coefficients
R0_TABLE = np.array(
    [
        [100013989.0, 0.0, 0.0],
        [1670700.0, 3.0984635, 6283.07585],
        [13956.0, 3.05525, 12566.1517],
        [3084.0, 5.1985, 77713.7715],
        [1628.0, 1.1739, 5753.3849],
        [1576.0, 2.8469, 7860.4194],
        [925.0, 5.453, 11506.77],
        [542.0, 4.564, 3930.21],
        [472.0, 3.661, 5884.927],
        [346.0, 0.964, 5507.553],
        [329.0, 5.9, 5223.694],
        [307.0, 0.299, 5573.143],
        [243.0, 4.273, 11790.629],
        [212.0, 5.847, 1577.344],
        [186.0, 5.022, 10977.079],
        [175.0, 3.012, 18849.228],
        [110.0, 5.055, 5486.778],
        [98.0, 0.89, 6069.78],
        [86.0, 5.69, 15720.84],
        [86.0, 1.27, 161000.69],
        [65.0, 0.27, 17260.15],
        [63.0, 0.92, 529.69],
        [57.0, 2.01, 83996.85],
        [56.0, 5.24, 71430.7],
        [49.0, 3.25, 2544.31],
        [47.0, 2.58, 775.52],
        [45.0, 5.54, 9437.76],
        [43.0, 6.01, 6275.96],
        [39.0, 5.36, 4694.0],
        [38.0, 2.39, 8827.39],
        [37.0, 0.83, 19651.05],
        [37.0, 4.9, 12139.55],
        [36.0, 1.67, 12036.46],
        [35.0, 1.84, 2942.46],
        [33.0, 0.24, 7084.9],
        [32.0, 0.18, 5088.63],
        [32.0, 1.78, 398.15],
        [28.0, 1.21, 6286.6],
        [28.0, 1.9, 6279.55],
        [26.0, 4.59, 10447.39],
    ]
)
R1_TABLE = np.array(
    [
        [103019.0, 1.10749, 6283.07585],
        [1721.0, 1.0644, 12566.1517],
        [702.0, 3.142, 0.0],
        [32.0, 1.02, 18849.23],
        [31.0, 2.84, 5507.55],
        [25.0, 1.32, 5223.69],
        [18.0, 1.42, 1577.34],
        [10.0, 5.91, 10977.08],
        [9.0, 1.42, 6275.96],
        [9.0, 0.27, 5486.78],
    ]
)
R2_TABLE = np.array(
    [
        [4359.0, 5.7846, 6283.0758],
        [124.0, 5.579, 12566.152],
        [12.0, 3.14, 0.0],
        [9.0, 3.63, 77713.77],
        [6.0, 1.87, 5573.14],
        [3.0, 5.47, 18849.23],
    ]
)
R3_TABLE = np.array([[145.0, 4.273, 6283.076], [7.0, 3.92, 12566.15]])
R4_TABLE = np.array([[4.0, 2.56, 6283.08]])

HELIO_R = [R0_TABLE, R1_TABLE, R2_TABLE, R3_TABLE, R4_TABLE]

# longitude and obliquity nutation coefficients
NUTATION_ABCD_ARRAY = np.array(
    [
        [-171996, -174.2, 92025, 8.9],
        [-13187, -1.6, 5736, -3.1],
        [-2274, -0.2, 977, -0.5],
        [2062, 0.2, -895, 0.5],
        [1426, -3.4, 54, -0.1],
        [712, 0.1, -7, 0],
        [-517, 1.2, 224, -0.6],
        [-386, -0.4, 200, 0],
        [-301, 0, 129, -0.1],
        [217, -0.5, -95, 0.3],
        [-158, 0, 0, 0],
        [129, 0.1, -70, 0],
        [123, 0, -53, 0],
        [63, 0, 0, 0],
        [63, 0.1, -33, 0],
        [-59, 0, 26, 0],
        [-58, -0.1, 32, 0],
        [-51, 0, 27, 0],
        [48, 0, 0, 0],
        [46, 0, -24, 0],
        [-38, 0, 16, 0],
        [-31, 0, 13, 0],
        [29, 0, 0, 0],
        [29, 0, -12, 0],
        [26, 0, 0, 0],
        [-22, 0, 0, 0],
        [21, 0, -10, 0],
        [17, -0.1, 0, 0],
        [16, 0, -8, 0],
        [-16, 0.1, 7, 0],
        [-15, 0, 9, 0],
        [-13, 0, 7, 0],
        [-12, 0, 6, 0],
        [11, 0, 0, 0],
        [-10, 0, 5, 0],
        [-8, 0, 3, 0],
        [7, 0, -3, 0],
        [-7, 0, 0, 0],
        [-7, 0, 3, 0],
        [-7, 0, 3, 0],
        [6, 0, 0, 0],
        [6, 0, -3, 0],
        [6, 0, -3, 0],
        [-6, 0, 3, 0],
        [-6, 0, 3, 0],
        [5, 0, 0, 0],
        [-5, 0, 3, 0],
        [-5, 0, 3, 0],
        [-5, 0, 3, 0],
        [4, 0, 0, 0],
        [4, 0, 0, 0],
        [4, 0, 0, 0],
        [-4, 0, 0, 0],
        [-4, 0, 0, 0],
        [-4, 0, 0, 0],
        [3, 0, 0, 0],
        [-3, 0, 0, 0],
        [-3, 0, 0, 0],
        [-3, 0, 0, 0],
        [-3, 0, 0, 0],
        [-3, 0, 0, 0],
        [-3, 0, 0, 0],
        [-3, 0, 0, 0],
    ]
)

NUTATION_YTERM_ARRAY = np.array(
    [
        [0, 0, 0, 0, 1],
        [-2, 0, 0, 2, 2],
        [0, 0, 0, 2, 2],
        [0, 0, 0, 0, 2],
        [0, 1, 0, 0, 0],
        [0, 0, 1, 0, 0],
        [-2, 1, 0, 2, 2],
        [0, 0, 0, 2, 1],
        [0, 0, 1, 2, 2],
        [-2, -1, 0, 2, 2],
        [-2, 0, 1, 0, 0],
        [-2, 0, 0, 2, 1],
        [0, 0, -1, 2, 2],
        [2, 0, 0, 0, 0],
        [0, 0, 1, 0, 1],
        [2, 0, -1, 2, 2],
        [0, 0, -1, 0, 1],
        [0, 0, 1, 2, 1],
        [-2, 0, 2, 0, 0],
        [0, 0, -2, 2, 1],
        [2, 0, 0, 2, 2],
        [0, 0, 2, 2, 2],
        [0, 0, 2, 0, 0],
        [-2, 0, 1, 2, 2],
        [0, 0, 0, 2, 0],
        [-2, 0, 0, 2, 0],
        [0, 0, -1, 2, 1],
        [0, 2, 0, 0, 0],
        [2, 0, -1, 0, 1],
        [-2, 2, 0, 2, 2],
        [0, 1, 0, 0, 1],
        [-2, 0, 1, 0, 1],
        [0, -1, 0, 0, 1],
        [0, 0, 2, -2, 0],
        [2, 0, -1, 2, 1],
        [2, 0, 1, 2, 2],
        [0, 1, 0, 2, 2],
        [-2, 1, 1, 0, 0],
        [0, -1, 0, 2, 2],
        [2, 0, 0, 2, 1],
        [2, 0, 1, 0, 0],
        [-2, 0, 2, 2, 2],
        [-2, 0, 1, 2, 1],
        [2, 0, -2, 0, 1],
        [2, 0, 0, 0, 1],
        [0, -1, 1, 0, 0],
        [-2, -1, 0, 2, 1],
        [-2, 0, 0, 0, 1],
        [0, 0, 2, 2, 1],
        [-2, 0, 2, 0, 1],
        [-2, 1, 0, 2, 1],
        [0, 0, 1, -2, 0],
        [-1, 0, 1, 0, 0],
        [-2, 1, 0, 0, 0],
        [1, 0, 0, 0, 0],
        [0, 0, 1, 2, 0],
        [0, 0, -2, 2, 2],
        [-1, -1, 1, 0, 0],
        [0, 1, 1, 0, 0],
        [0, -1, 1, 2, 2],
        [2, -1, -1, 2, 2],
        [0, 0, 3, 2, 2],
        [2, -1, 0, 2, 2],
    ]
)
# END COEFF

# now, we write to the actual code

DELTA_T = 67


@jit
def helio_vector(vec, jme):
    """This function calculates equation 9 across the vector"""
    return np.sum(
        vec[:, 0] * np.cos(vec[:, 1] + vec[:, 2] * jme[..., np.newaxis]), axis=-1
    )


def solar_position(timestamp, latitude, longitude, elevation):
    """Calculate the position of the sun at a given location and time.

    Args:
        timestamp (array-like): The timestamp(s) at each point.
        latitude (array-like): The latitude(s) of each point.
        longitude (array-like): The longitude(s) of each point.
        elevation (array-like): The elevation of each point.

    Returns:
        ndarray: An array containing the altitude and azimuth for each point.
    """
    jd = timestamp / 86400.0 + 2440587.5
    jc = (jd - 2451545) / 36525

    jde = jd + DELTA_T / 86400.0
    jce = (jde - 2451545) / 36525

    jm = jc / 10
    jme = jce / 10

    # todo: make more elegant?
    # TODO: vectorize later? it's kinda complex

    # heliocentric longitude
    l_rad = np.zeros_like(timestamp)
    for idx, vec in enumerate(HELIO_L):
        l_rad = l_rad + helio_vector(vec, jme) * jme ** idx

    l_rad = l_rad / 10e8
    l_deg = np.rad2deg(l_rad) % 360

    # heliocentric latitude
    b_rad = np.zeros_like(timestamp)
    for idx, vec in enumerate(HELIO_B):
        b_rad = b_rad + helio_vector(vec, jme) * jme ** idx
    b_rad = b_rad / 10e8
    b_deg = np.rad2deg(b_rad) % 360

    # heliocentric radius
    r_rad = np.zeros_like(timestamp)
    for idx, vec in enumerate(HELIO_R):
        r_rad = r_rad + helio_vector(vec, jme) * jme ** idx
    r_rad = r_rad / 10e8
    r_deg = np.rad2deg(r_rad) % 360

    theta = (l_deg + 180) % 360
    beta = -1 * b_deg

    def cubic_poly(a, b, c, d):
        return a + b * jce + c * jce ** 2 + (jce ** 3) / d

    X0 = cubic_poly(297.85036, 445267.111480, -0.0019142, 189474)
    X1 = cubic_poly(357.52772, 35999.050340, -0.0001603, -300000)
    X2 = cubic_poly(134.96298, 477198.867398, 0.0086972, 56250)
    X3 = cubic_poly(93.27191, 483202.017538, -0.0036825, 327270)
    X4 = cubic_poly(125.04452, 1934.136261, 0.0020708, 450000)

    X = np.vstack([X0, X1, X2, X3, X4]).T

    nut = NUTATION_ABCD_ARRAY

    # TODO: these are gross - use loops instead of broadcasting?
    # FIXME: use guvectorize, treat jce as a scalar.
    d_psi = (nut[:, 0] + jce[..., np.newaxis] * nut[:, 1]) * np.sin(
        np.sum(X[:, np.newaxis, :] * NUTATION_YTERM_ARRAY[np.newaxis, ...], axis=2)
    )
    d_psi = np.sum(d_psi, axis=-1) / 36, 000, 000

    d_epsilon = (nut[:, 2] + jce[..., np.newaxis] * nut[:, 3]) * np.cos(
        np.sum(X[:, np.newaxis, :] * NUTATION_YTERM_ARRAY[np.newaxis, ...], axis=2)
    )
    d_epsilon = np.sum(d_epsilon, axis=-1) / 36, 000, 000

    u = jme[:, np.newaxis] / 10 * np.arange(0, 10).reshape((1, -1))
    epsilon_0 = np.array(
        [
            84381.448,
            -4680.93,
            1.55,
            1999.25,
            -51.38,
            -249.67,
            -39.05,
            7.12,
            27.87,
            5.79,
            2.45,
        ]
    )
    epsilon = np.sum(u * epsilon_0, axis=-1) / 3600 + d_epsilon
    d_tau = -20.4898 / (3600 * r_deg)
    sun_longitude = theta + d_psi + d_tau

    v_0 = 280.46061837 + 360.98564736629 * (jd - 2451545) + 0.000387933 * jc ** 2 - jc ** 3 / 38710000
    v_0 = v_0 % 360

    v = v_0 + d_psi * np.cos(np.deg2rad(epsilon))

    alpha = np.arctan2(np.sin(np.radians(sun_longitude)) *
                       np.cos(np.radians(epsilon)) -
                       np.tan(np.radians(beta)) *
                       np.sin(np.radians(epsilon)),
                       np.cos(np.radians(sun_longitude)))
    alpha_deg = np.rad2deg(alpha) % 360
    delta = np.arcsin(
        np.sin(np.radians(beta)) *
        np.cos(np.radians(epsilon)) +
        np.cos(np.radians(beta)) *
        np.sin(np.radians(epsilon)) *
        np.cos(np.radians(sun_longitude))
    )
    delta_deg = np.rad2deg(delta) % 360

    h = v + latitude - alpha_deg

    xi_deg = 8.794 / (3600 * r_deg)
    u = np.arctan(0.99664719 * np.tan(latitude))

    x = np.cos(u) + elevation / 6378140 * np.cos(latitude)

    y = 0.99664719 * np.sin(u) + elevation / 6378140 * np.sin(latitude)

    d_alpha = np.arctan2(-1 * x * np.sin(np.radians(xi_deg)) * np.sin(np.radians(h)), np.cos(delta))
    d_alpha = np.rad2deg(d_alpha)
    alpha_prime = alpha_deg + d_alpha
    delta_prime = np.arctan2((np.sin(delta) - y * np.sin(np.radians(xi_deg))) * np.cos(np.radians(d_alpha)),
                             np.cos(delta) - x * np.sin(np.radians(xi_deg)) * np.cos(np.radians(h)))
    topo_local_hour_angle_deg = h - d_alpha
