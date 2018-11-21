import ray
from typing import List

from server.transaction import Transaction

@ray.remote
class Worker(object):
	def __init__(self, depth: int, worker_id: int, parent=None):
		self.children = None
		self.depth = depth
		self.worker_id = worker_id
		self.parent = parent
		self.write_transaction_list = list()
		self.read_transaction_list = list()

		if parent:
			parent_depth = ray.get(parent.get_depth.remote())
			prefix = format(worker_id, '0{}b'.format(parent_depth))
			self.start = prefix + '0' * depth
			self.end = prefix + '1' * depth
			print(worker_id, self.start, self.end, len(self.start), len(self.end))
			print()


	def set_children(self, children) -> None:
		self.children = children

	def get_depth(self) -> int:
		return self.depth

	def receive_transaction(self, transaction: Transaction) -> None:
		if transaction.transaction_type == 'W':
			write_transaction_list.append(transaction)
		else:
			read_transaction_list.append(transaction)

	# TODO: batch update function
	# def batch():
	# 	vals = []
	# 	for child self.children: 
	# 		vals += ray.get(child.batch.remote())
