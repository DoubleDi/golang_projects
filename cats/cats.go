package cats

/*
There is a special breed of cat, and you are helping with organising a beauty contest for them.
A cat from this breed has the exact same length of fur (down to the micrometer) all over it.
There are N cats in the contest, and K judges scoring them.
As there are too many cats for each judge to look at, the judges just give their preferred fur length (given in an array).
You also have the fur length for each of the cats (also given in an array).
The scoring is the following:
 - Each cat gets penalty points from each judge for every micrometer difference from their preferred fur length.
   - E.g.: If the preferred length is 200 micrometers, and the fur of the cat is 1000 micrometers long, the cat gets 800 penalty points from that judge.
 - These penalty points for a given cat adds together from all of the judges.
 - The cat with the least penalty points wins.

 The task is to find the winning cat from the input data.

Example:

N = 4, K = 4
fur_length = [2, 0, 8, 4]
judges = [2, 7, 9, 1]


fur_length         f_2       f_1       f_4                 f_3
                    |         |         |                   |
                    0----1----2----3----4----5----6----7----8---9----> in micrometers
                         |    |                        |        |
judge_preference        j_4  j_1                      j_2      j_3

Cumulative penalty points for f_4: 2 + 3 + 5 + 3 = 13
                                   |   |   |   |
                                  j_1 j_2 j_3 j_4

double loop iteration O(N * K)

2
2 7 9 1
2 - 2 + abs(2 - 7) + abs(2 - 9) + abs(2 - 1)

2 - 2 + 2 - 1  = cat_len * len(score smaller) - sum(score smaller)
(7-2) + (9 - 2) = sum(score bigger) - cat_len * len(score bigger)

soriting: O(K * log(K))
binary search for each cat: O(N * log(K))
sum calculation: O(K)
O((n+k) * log(K))

N = 1 K = 4
N * K = 4
5 * 2 = 10
*/

func WinnerCat(cats, judges []int) (int, int) {
	if len(cats) == 0 || len(judges) == 0 {
		return 0, -1
	}
	m := -1
	ind := -1
	for i, c := range cats {
		cum := 0
		for _, j := range judges {
			if c < j {
				cum += j - c
			} else {
				cum += c - j
			}
		}
		if m == -1 || cum < m {
			m = cum
			ind = i
		}
	}
	return m, ind
}
