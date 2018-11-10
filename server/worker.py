import ray
from typing import List

@ray.remote
class Worker(object):
	def __init__(self, parent: Worker=None, depth: int, worker_id: int):
		self.parent = parent
		self.children = None
		self.depth = depth
		self.worker_id = worker_id

		prefix = format(worker_id, '0{}b'.format(parent.depth))
		self.start = prefix + '0' * depth
		self.end = prefix + '1' * depth

	def set_children(self, children: List[Worker]) -> None:
		self.children = children

