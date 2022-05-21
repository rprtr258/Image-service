#include <stdint.h>
#include <immintrin.h>
#include <string.h>

static __m256i my_mm256_abs_epi64(__m256i x) {
	// i64 mask = x >> 63;
    __m256i mask = _mm256_cmpgt_epi64(_mm256_srli_epi64(x, 63), _mm256_set_epi64x(0, 0, 0, 0));
	// return (x + mask) ^ mask;
    return _mm256_xor_si256(_mm256_add_epi64(x, mask), mask);
}

int64_t fastDist(int64_t* a, int64_t* b) {
    __m256i ma = _mm256_loadu_si256((__m256i*)a);
    __m256i mb = _mm256_loadu_si256((__m256i*)b);
    __m256i ma_sub_b = _mm256_sub_epi64(mb, ma);
    __m256i mabs_a_sub_b = my_mm256_abs_epi64(ma_sub_b);
    int64_t r[4];
    _mm256_storeu_si256((__m256i*)&r, mabs_a_sub_b);
    return r[0] + r[1] + r[2];
}

static size_t ssqrt(size_t x) {
    size_t res = 0;
    while (res * res <= x) {
        ++res;
    }
    return res - 1;
}

void kmeansIters(int64_t clustersCenters[][3], size_t clustersCentersLen, int64_t pixelColors[][3], size_t pixelColorsLen) {
	size_t batchMaxSize = ssqrt(pixelColorsLen);
    size_t sumAndCountLen = sizeof(int64_t) * clustersCentersLen * 4;
	int64_t* sumAndCount = malloc(sumAndCountLen); // count and sum of Rs, Gs, Bs
	for (int epoch = 0; epoch < 300; ++epoch) {
		size_t k = (size_t)rand() % (batchMaxSize + 1);
        memset(sumAndCount, 0, sumAndCountLen);
		for (size_t i = k; i < pixelColorsLen; i += k) {
			int64_t* pixelColor = pixelColors[i];
			int minCluster = 0;
			int64_t minDist = fastDist(pixelColor, clustersCenters[0]);
			for (int k = 1; k < clustersCentersLen; ++k) {
				int64_t newDist = fastDist(pixelColor, clustersCenters[k]);
				if (newDist < minDist) {
					minCluster = k;
					minDist = newDist;
				}
			}
			sumAndCount[minCluster*4+0] += 1;
			sumAndCount[minCluster*4+1] += pixelColor[0];
			sumAndCount[minCluster*4+2] += pixelColor[1];
			sumAndCount[minCluster*4+3] += pixelColor[2];
		}
		int64_t movement = 0;
		for (size_t i = 0; i < clustersCentersLen; ++i) {
			int64_t count = sumAndCount[i*4+0];
			if (count == 0) {
				continue;
			}
			sumAndCount[i*4+1] /= count;
			sumAndCount[i*4+2] /= count;
			sumAndCount[i*4+3] /= count;
			movement += fastDist(clustersCenters[i], &sumAndCount[i*4+1]);
            memcpy(clustersCenters[i], &sumAndCount[i*4+1], 3*sizeof(int64_t));
		}
		if (movement < 100) {
			break;
		}
	}
    free(sumAndCount);
}