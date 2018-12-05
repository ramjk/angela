import bisect
import ray

from server import smt_api 
from typing import List, Optional, Tuple
from server.transaction import Transaction, WriteTransaction
from common.util import random_string, to_bytes, to_string

@ray.remote
class Worker(object):
	def __init__(self, depth: int, worker_id: int, prefix_length: int, parent=None):
		self.children = None
		self.depth = depth
		self.worker_id = worker_id
		self.parent = parent
		self.write_transaction_list = list()
		self.read_transaction_list = list()
		self.prefix = ""

		if parent:
			self.prefix = format(worker_id, '0{}b'.format(prefix_length))
			self.start = self.prefix + '0' * depth
			self.end = self.prefix + '1' * depth

	def set_children(self, children) -> None:
		self.children = children

	def get_depth(self) -> int:
		return self.depth

	def receive_transaction(self, transaction: Transaction) -> Optional[str]:
		# print("In worker {}".format(self.worker_id))
		# print("Transaction ID: {}".format(transaction.Index))
		if transaction.TransactionType == 'W':
			bisect.insort(self.write_transaction_list, transaction)
		else:
			return smt_api.read(transaction.Index)
			
	def batch_update(self, epoch_number, worker_roots=None):
		print("performing batch update for workerID:", self.worker_id)
		if worker_roots:
			transaction_list = list()
			for i in range(0, len(worker_roots), 2):
				worker_root_1 = bytearray(to_bytes(worker_roots[i][1]))
				worker_root_2 = bytearray(to_bytes(worker_roots[i+1][1]))
				worker_root_1.extend(worker_root_2)
				transaction_digest = to_string(bytes(worker_root_1))
				transaction_index = worker_roots[i][0][:-1]
				# ^ remove last bit because that is the common prefix between the i and i+1 nodes
				transaction_list.append(WriteTransaction(transaction_index, transaction_digest))
		else:
			transaction_list = self.write_transaction_list

		keys = list() 
		values = list()
		
		for transaction in transaction_list:
			keys.append(transaction.Index[len(self.prefix):])
			values.append(transaction.Data)

		# worker_root_digest = random_string()
		worker_root_digest = smt_api.batch_insert(self.prefix, keys, values, epoch_number)
		return worker_root_digest, self.prefix

