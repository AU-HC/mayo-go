//go:build mayo1

package mayo

const N = 86
const M = 78
const mVecLimbs = 5
const o = 8
const v = N - o
const aCols = K*o + 1
const K = 10
const q = 16
const OBytes = 312
const vBytes = 39
const rBytes = 40
const P1Bytes = 120159
const P2Bytes = 24336
const P3Bytes = 1404
const cskBytes = 24
const cpkBytes = 1420
const eskBytes = 144831
const epkBytes = 145899
const lBytes = 24336
const sigBytes = 454
const saltBytes = 24
const digestBytes = 32
const pkSeedBytes = 16
const skSeedBytes = 24
const tailFLength = 4

const P1Limbs = v * (v + 1) / 2 * mVecLimbs
const P2Limbs = v * o * mVecLimbs
const P3Limbs = o * (o + 1) / 2 * mVecLimbs

var tailF = []byte{8, 1, 1, 0}
