import socket
import signal
import argparse
import ray

from math import log

from server.transaction import Transaction
from server.worker import Worker

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

		ray.init()

		self.root_depth = int(log(num_workers-1, 2))
		self.root_worker = Worker.remote(self.root_depth, -1)
		self.leaf_workers = leaf_workers = list()
		for worker_id in range(num_workers-1):
			worker = Worker.remote(tree_depth-self.root_depth, worker_id, self.root_worker)
			leaf_workers.append(worker)
		self.root_worker.set_children.remote(leaf_workers)

	def start(self) -> None:
		try:
			while True:
				self.receive()

		except KeyboardInterrupt as err:
			if new_socket:
				new_socket.close()
			self.close()

	def receive(self) -> None:
		(new_socket, address) = self.socket.accept()
		try:
			msg = new_socket.recv(1024)
			if msg == b"practice":
				print(msg)
				new_socket.send(msg)
		finally:
			new_socket.close()

	def close(self) -> None:
		print("Closing server...")
		self.socket.close()
		ray.shutdown()

	def receive_transaction(self, transaction: Transaction) -> None:
		# destination_worker_index = int(transaction.index[:int(log(self.num_workers, 2))], 2) # FIXME: Can we use a more efficient representation of indices/node_ids?
		destination_worker_index = int(transaction.index, 2) >> (len(transaction.index) - self.root_depth)
		ray.get(self.leaf_workers[destination_worker_index].receive_transaction.remote(transaction))
			
if __name__ == '__main__':
	parser = argparse.ArgumentParser()
	parser.add_argument('port', type=int, help="port number")
	parser.add_argument('num_workers', type=int, help="number of worker nodes")
	parser.add_argument('epoch_length', type=int, help="number of transactions per epoch")
	parser.add_argument('tree_depth', type=int, help="depth of tree")
	args = parser.parse_args()

	server = Server(args.port, args.num_workers, args.epoch_length, args.tree_depth)

	server.start()
