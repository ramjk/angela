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

		root = Client.get_signed_root()
		

		# self.server = Server(port, num_worker, epoch_length, tree_depth)
		# self.client = Client("localhost", port)

		# for i in range(4):
		# 	print()
		# print("__________________________________________")

	# def tearDown(self):
		# self.server.close()
		# self.client.end_session()

	def test_practice(self):
		index = util.random_index()
		data = util.random_string()
		for i in range(1):
			index = util.random_index()
			data = util.random_string()
			tx = Client.insert_leaf(index, data)
		# tx = Client.insert_leaf(index, data)
		# root = Client.get_signed_root()
		# proof = Client.generate_proof(index)
		# self.assertTrue(Client.verify_proof(proof, data, root))

		# index = util.random_index()
		# data = util.random_string()
		# root = Client.get_signed_root()
		# self.assertTrue(Client.verify_proof(Client.generate_proof(index), data, root))

if __name__ == '__main__':
	unittest.main()
