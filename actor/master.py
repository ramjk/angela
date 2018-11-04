import ray, time
from common import util

@ray.remote
class Master(object):

	def __init__(self, transactions: dict):
		self.leaves = sorted(transactions.items(), key=lambda leaf:leaf[0])
		self.conflicts = util.find_conflicts(list(self.leaves.keys()))

