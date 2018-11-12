import socket
import argparse
import ray
from math import log

from transaction import Transaction
from worker import Worker

class Server(object):
	"""
	Assumptions:
	    - N = num_workers where N - 1 is a power of two.
	"""
	def __init__(self, port: int, num_workers: int, epoch_length: int):
		self.port = port
		self.num_workers = num_workers
		self.epoch_length = epoch_length
		self.socket = socket.socket()
		self.socket.bind(('', port))
		self.socket.listen()

		root_depth = int(log(num_workers-1, 2))
		self.root_worker = Worker.remote(root_depth, -1)
		self.leaf_workers = leaf_workers = list()
		for worker_id in range(num_workers-1):
			worker = Worker.remote(3-root_depth, worker_id, self.root_worker)
			leaf_workers.append(worker)
		self.root_worker.set_children.remote(leaf_workers)


	def start(self) -> None:
		while True:
			(new_socket, address) = self.socket.accept()
			msg = new_socket.recv(1024)
			tmp = msg
			while tmp:
				tmp = new_socket.recv(1024)
				msg += tmp
			print(msg)


	def receive_transaction(self, transaction: Transaction) -> None:
		destination_worker_index = int(transaction.index, 2) / (self.num_workers - 1)
		self.leaf_workers[destination_worker_index].receive_transaction(transaction)
			

parser = argparse.ArgumentParser()
parser.add_argument('port', type=int, help="port number")
parser.add_argument('num_workers', type=int, help="number of worker nodes")
parser.add_argument('epoch_length', type=int, help="number of transactions per epoch")
args = parser.parse_args()

ray.init()

server = Server(args.port, args.num_workers, args.epoch_length)


