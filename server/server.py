import socket
import argparse
import ray
from math import log

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

		root_depth = log(num_workers - 1, 2)
		self.root_worker = Worker.remote(root_depth)
		self.leaf_workers = leaf_workers = list()
		for worker_id in range(root_depth):
			worker = Worker.remote(self.root_worker, 128 - root_depth, worker_id)
			leaf_workers.append(worker)
		self.root_worker.set_children(leaf_workers)


	def start(self) -> None:
		while True:
			(new_socket, address) = self.socket.accept()
			msg = new_socket.recv(1024)
			tmp = msg
			while tmp:
				tmp = new_socket.recv(1024)
				msg += tmp
			print(msg)

parser = argparse.ArgumentParser()
parser.add_argument('port', type=int, help="port number")
parser.add_argument('num_workers', type=int, help="number of worker nodes")
parser.add_argument('epoch_length', type=int, help="number of transactions per epoch")
args = parser.parse_args()

server = Server(args.port, args.num_workers, args.epoch_length)

ray.init()

