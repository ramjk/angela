import matplotlib.pyplot as plt
import merkle.smt
import timeit
from common import util

durations = list()
num_inserts = list()
for i in range(6, 11):
	durations.append(timeit.timeit('T.insert(util.random_index(), util.random_string())', setup='import merkle.smt; from common import util; T = merkle.smt.SparseMerkleTree("sha256")', number=2**i))
	num_inserts.append(2**i)
plt.plot(num_inserts, durations, "bo")
plt.ylabel('Duration of random insert workload')
plt.xlabel('Number of random inserts')
plt.xticks([i for i in num_inserts], num_inserts)
plt.show()