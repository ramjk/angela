import socket
import signal
import argparse
import ray
import json

from flask import Flask
from multiprocessing import Process
from math import log

from server.smt_api import getLatestRootDigest
from server.transaction import Transaction, WriteTransaction
from server.worker import Worker
from common.util import random_index, random_string, send_data

app = Flask(__name__)

class Server(object):
	"""
	Assumptions:
	    - N = num_workers where N - 1 is a power of two.
	"""
	def __init__(self, port: int, num_workers: int, epoch_length: int, tree_depth: int, flag: bool):
		self.port = port
		self.num_workers = num_workers
		self.epoch_length = epoch_length
		self.tree_depth = tree_depth
		self.socket = socket.socket()
		self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
		self.socket.bind(("localhost", port))
		self.socket.listen()

		self.epoch_number = 1

		self.write_transaction_list = list()
		self.read_transaction_list = list()

		if flag:
			ray.init()

		# num_workers includes the root node
		self.prefix_length = int(log(num_workers-1, 2))
		self.root_worker = Worker.remote(self.prefix_length-1, -1, self.prefix_length)
		self.leaf_workers = leaf_workers = list()
		for worker_id in range(num_workers-1):
			worker = Worker.remote(tree_depth-self.prefix_length, worker_id, self.prefix_length, self.root_worker)
			leaf_workers.append(worker)
		self.root_worker.set_children.remote(leaf_workers)

	def start(self) -> None:
		# pool = Pool()
		try:
			while True:
				(conn, address) = self.socket.accept()
				process = Process(target=self.receive, args=(conn,))
				process.daemon = True
				process.start()
				# print("Async apply")
				# pool.apply_async(self.receive, (conn,))
				# self.receive(conn)
		except KeyboardInterrupt as err:
			pool.terminate()
			self.close()

	def receive(self, conn) -> None:
		print("[receive]")
		data_length = int(conn.recv(200))
		print("[receive] 65")
		data_type = str(conn.recv(200))

		print("[receive] 68")
		tmp = msg = conn.recv(200)	
		data_length -= len(tmp)	

		print("[receive] 72")
		while data_length > 0:	
			tmp = conn.recv(200)
			msg += tmp
			data_length -= len(tmp)

		print("[receive] 78")

		tx = Transaction.from_dict(json.loads(msg))

		print("[receive] 82")
		if tx.TransactionType == "W" and tx.Data == "practice":
			send_data(conn, tx)

		result = self.receive_transaction(tx)
		print("Sending data")
		send_data(conn, result)
		print("[receive] 89")
		conn.close()

	def close(self) -> None:
		print("Closing server...")
		self.socket.close()
		map(lambda worker: ray.shutdown(worker), self.leaf_workers)
		ray.shutdown()

	def receive_transaction(self, transaction: Transaction) -> None:
		print("99")
		print(transaction.__dict__)
		if transaction.TransactionType == "R" and transaction.Index == "":
			return getLatestRootDigest()
		# destination_worker_index = int(transaction.Index[:int(log(self.num_workers, 2))], 2) # FIXME: Can we use a more efficient representation of indices/node_ids?
		destination_worker_index = int(transaction.Index, 2) >> (len(transaction.Index) - self.prefix_length)
		print(104)
		object_id = self.leaf_workers[destination_worker_index].receive_transaction.remote(transaction)
		print("106")

		if transaction.TransactionType == 'W':
			self.write_transaction_list.append(transaction)
		else:
			print("110")
			return ray.get(object_id) # If read, then return proof object

		worker_roots = list()
		object_ids = list()

		if len(self.write_transaction_list) == self.epoch_length:
			for leaf_worker in self.leaf_workers:
				object_ids.append(leaf_worker.batch_update.remote(self.epoch_number))

			for object_id in object_ids:
				worker_root_digest, prefix = ray.get(object_id)
				worker_roots.append((prefix, worker_root_digest))

			ray.get(self.root_worker.batch_update.remote(self.epoch_number, worker_roots))
			self.write_transaction_list = list()
			self.epoch_number += 1

		return "True"
			
# if __name__ == '__main__':
	# parser = argparse.ArgumentParser()
	# parser.add_argument('port', type=int, help="port number")
	# parser.add_argument('num_workers', type=int, help="number of worker nodes")
	# parser.add_argument('epoch_length', type=int, help="number of transactions per epoch")
	# parser.add_argument('tree_depth', type=int, help="depth of tree")
	# args = parser.parse_args()

	# server = Server(args.port, args.num_workers, args.epoch_length, args.tree_depth)
	# server.start()

	# server = Server(8008, 9, 1000, 256)
	# for i in range(1000):
	# 	transaction = WriteTransaction(random_index(), random_string())
	# 	server.receive_transaction(transaction)
