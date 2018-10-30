import matplotlib.pyplot as plt
import merkle.smt
import timeit
from common import util

durations = list()
num_inserts = list()
setup = '''
import merkle.smt
from common import util

T = merkle.smt.SparseMerkleTree("sha256")
'''
stmt = "T.insert(util.random_index(digest_size=merkle.smt.SparseMerkleTree.TREE_DEPTH), util.random_string())"

for i in range(6, 14):
	durations.append(timeit.timeit(stmt=stmt, setup=setup, number=2**i))
	num_inserts.append(2**i)
plt.plot(num_inserts, durations, "bo")
plt.ylabel('Duration of random insert workload')
plt.xlabel('Number of random inserts')
plt.xticks([i for i in num_inserts], num_inserts)
plt.title("Durations of Single, Random Inserts")
plt.show()
print("durations from first run", durations)

durations = list()
num_inserts = list()
batch_setup = '''
import merkle.smt
from common import util

T = merkle.smt.SparseMerkleTree("sha256")
num = {}
values = [util.random_string() for _ in range(num)]
leaves = [util.random_index(digest_size=merkle.smt.SparseMerkleTree.TREE_DEPTH) for _ in range(num)]
d = {{k:v for (k,v) in zip(leaves, values)}}
'''
batch_stmt = "T.batch_insert(d)"
for i in range(6, 14):
	durations.append(timeit.timeit(stmt=batch_stmt, setup=batch_setup.format(2**i), number=1))
	num_inserts.append(2**i)
plt.plot(num_inserts, durations, "bo")
plt.ylabel('Duration of random batch insert workload')
plt.xlabel('Number of random batch inserts')
plt.xticks([i for i in num_inserts], num_inserts)
plt.title("Durations of Batch, Random Inserts")
plt.show()

print("durations from batch run", durations)
