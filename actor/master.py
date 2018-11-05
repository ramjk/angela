import ray, time
from common import util

@ray.remote
class Master(object):

	def __init__(self, transactions: dict):
		self.leaves = sorted(transactions.items(), key=lambda leaf:leaf[0])
		self.conflicts = util.find_conflicts(list(self.leaves.keys()))


@ray.remote
def find_conflicts_batch(leaves):
	result = []
	for i in range(1, len(leaves)):
		x, y = util.bitarray(leaves[i-1]), util.bitarray(leaves[i])
		z = x^y
		k = len(x)
		for idx, elem in enumerate(z):
			if elem:
				k = idx
				break
		result.append(x[0:k].to01())
	return result

def find_conflicts_serial(leaves):
	result = []
	for i in range(1, len(leaves)-3):
		x, y = util.bitarray(leaves[i-1]), util.bitarray(leaves[i])
		z = x^y
		k = len(x)
		for idx, elem in enumerate(z):
			if elem:
				k = idx
				break
		result.append(x[0:k].to01())
	return result

ray.init()
print("Parallel")
vectors = [util.random_index(digest_size=256) for i in range(80000)]
vectors.sort()
batches = [vectors[i:i+20000] for i in range(0, len(vectors) - 20000 + 1, 20000)]

start_time = time.time()

results = ray.get([find_conflicts_batch.remote(batches[i]) for i in range(0, len(batches))])

end_time = time.time()
duration = end_time - start_time
flattened = set([val for result in results for val in result])
print(duration, "seconds", len(flattened))
ray.shutdown()

print("Serial")
start_time = time.time()

results = find_conflicts_serial(vectors)

end_time = time.time()
duration = end_time - start_time
print(duration, "seconds", len(results))