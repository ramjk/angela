import matplotlib.pyplot as plt
import merkle.smt
import timeit
from common import util

durations = list()
iterations = list()
for i in range(11):
	durations.append(timeit.timeit('T.insert(util.random_index(), util.random_string())', setup='import merkle.smt; from common import util; T = merkle.smt.SparseMerkleTree("sha256")', number=2**i))
	iterations.append(2**i)
plt.plot(iterations, durations, "bo")
plt.show()