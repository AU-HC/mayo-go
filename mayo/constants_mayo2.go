//go:build mayo2

package mayo

const N = 81
const M = 64
const mVecLimbs = 4
const o = 17
const v = N - o
const aCols = K*o + 1
const K = 4
const q = 16
const OBytes = 544
const vBytes = 32
const rBytes = 24
const P1Bytes = 66560
const P2Bytes = 34816
const P3Bytes = 4896
const cskBytes = 24
const cpkBytes = 4912
const eskBytes = 101944
const epkBytes = 106272
const lBytes = 34816
const sigBytes = 186
const saltBytes = 24
const digestBytes = 32
const pkSeedBytes = 16
const skSeedBytes = 24
const tailFLength = 4

const P1Limbs = v * (v + 1) / 2 * mVecLimbs
const P2Limbs = v * o * mVecLimbs
const P3Limbs = o * (o + 1) / 2 * mVecLimbs

var tailF = [2]byte{8, 0, 2, 8}
