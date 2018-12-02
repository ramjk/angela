import bisect
import ray
from typing import List, Optional, Tuple

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

	def set_children(self, children) -> None:
		self.children = children

	def get_depth(self) -> int:
		return self.depth

	def receive_transaction(self, transaction: Transaction) -> None:
		print("In worker {}".format(self.worker_id))
		print("Transaction ID: {}".format(transaction.index))
		if transaction.transaction_type == 'W':
			bisect.insort(self.write_transaction_list, transaction)
		else:
			self.read_transaction_list.append(transaction)

	def batch_update(self, transaction_list=None, dirty_list=None):
		print("performing batch update for workerID:", self.worker_id)
		if not transaction_list:
			# call the c extension code here to perform batch percolate
			new_root = ('101', "chocolate")
			change_list = [('1001', 'x'), ('1000', 'y'), ('10010', 'z')]
			return new_root, change_list 

		else:
			print("===ROOT===")
			# call the c extension code here to perform batch 
			# percolate on root node
