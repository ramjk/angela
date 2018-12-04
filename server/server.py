import socket
import signal
import argparse
import ray
import json

from multiprocessing import Pool
from math import log
from server.transaction import Transaction
from server.worker import Worker
from common.util import send_data

class Server(object):
	"""
	Assumptions:
	    - N = num_workers where N - 1 is a power of two.
	"""
	def __init__(self, port: int, num_workers: int, epoch_length: int, tree_depth: int):
		self.port = port
		self.num_workers = num_workers
		self.epoch_length = epoch_length
		self.tree_depth = tree_depth
		self.socket = socket.socket()
		self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
		self.socket.bind(("localhost", port))
		self.socket.listen()

		self.write_transaction_list = list()
		self.read_transaction_list = list()
		self.in_write_mode = False

		ray.init()

		self.root_depth = int(log(num_workers-1, 2))
		self.root_worker = Worker.remote(self.root_depth, -1)
		self.leaf_workers = leaf_workers = list()
		for worker_id in range(num_workers-1):
			worker = Worker.remote(tree_depth-self.root_depth, worker_id, self.root_worker)
			leaf_workers.append(worker)
		self.root_worker.set_children.remote(leaf_workers)

	def start(self) -> None:
		pool = Pool()
		try:
			while True:
				(conn, address) = self.socket.accept()
				pool.apply_async(self.receive, (conn,))
		except KeyboardInterrupt as err:
			self.close()

	def receive(self, conn) -> None:
		data_length = int(conn.recv(200))
		data_type = str(conn.recv(200))

		print(data_length)

		tmp = msg = conn.recv(200)	
		data_length -= len(tmp)	

		while data_length > 0:	
			tmp = conn.recv(200)
			msg += tmp
			data_length -= len(tmp)

		tx = Transaction.from_dict(json.loads(msg))

		if tx.data == "practice":
			send_data(conn, tx)

		# self.receive_transaction(tx)
		conn.close()

	def close(self) -> None:
		print("Closing server...")
		self.socket.close()
		ray.shutdown()

	def receive_transaction(self, transaction: Transaction) -> None:
		if transaction.transaction_type == 'W':
			self.write_transaction_list.append(transaction)
		else:
			self.read_transaction_list.append(transaction)

		# destination_worker_index = int(transaction.index[:int(log(self.num_workers, 2))], 2) # FIXME: Can we use a more efficient representation of indices/node_ids?
		destination_worker_index = int(transaction.index, 2) >> (len(transaction.index) - self.root_depth)
		self.leaf_workers[destination_worker_index].receive_transaction.remote(transaction)

		new_roots = list()
		dirty_list = list()
		object_ids = list()

		if len(self.write_transaction_list) == self.epoch_length:
			for leaf_worker in self.leaf_workers:
				object_ids.append(leaf_worker.batch_update.remote())
				# call c-extension code here

			for object_id in object_ids:
				new_root, change_list = ray.get(object_id)
				new_roots.append(new_root)
				dirty_list.extend(change_list)

			print(new_roots)
			print(dirty_list)

			ray.get(self.root_worker.batch_update.remote(new_roots, dirty_list))

			self.write_transaction_list = list()
			
if __name__ == '__main__':
	parser = argparse.ArgumentParser()
	parser.add_argument('port', type=int, help="port number")
	parser.add_argument('num_workers', type=int, help="number of worker nodes")
	parser.add_argument('epoch_length', type=int, help="number of transactions per epoch")
	parser.add_argument('tree_depth', type=int, help="depth of tree")
	args = parser.parse_args()

	server = Server(args.port, args.num_workers, args.epoch_length, args.tree_depth)
	server.start()
