import unittest
import threading
import ray

from common import util
from server.transaction import WriteTransaction
from client.client import Client
from server.server import Server

class TestSparseMerkleTree(unittest.TestCase):
	def setUp(self):
		port = 3307
		num_worker = 9
		epoch_length = 1000
		tree_depth = 256

		self.server = Server(port, num_worker, epoch_length, tree_depth)
		self.client = Client("localhost", port)

		for i in range(4):
			print()
		print("__________________________________________")

		self.server_thread = threading.Thread(target=self.server.receive)

	def tearDown(self):
		self.server.close()
		self.client.end_session()

	def test_practice(self):
		root = Client.get_signed_root()
		print(root)
		index = util.random_index()
		data = util.random_string()
		tx = Client.insert_leaf(index, data)
		print(tx)
		root = Client.get_signed_root()
		proof = Client.generate_proof(index)
		print(Client.verify_proof(proof, data, root))

		index = util.random_index()
		data = util.random_string()
		root = Client.get_signed_root()
		print(Client.verify_proof(Client.generate_proof(index), data, root))

		# self.server.epoch_length = 2
		# for i in range(2):
		# 	print(self.client.insert_leaf(util.random_index(), util.random_string()))


		# print("36")
		# self.server.receive_transaction(WriteTransaction('1000', 'apple'))
		# print("38")
		# self.server.receive_transaction(WriteTransaction('1001', 'banana'))

if __name__ == '__main__':
	unittest.main()
