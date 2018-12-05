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
		epoch_length = 1
		tree_depth = 256

		# root = Client.get_signed_root()
	
		self.server = Server(port, num_worker, epoch_length, tree_depth)
		self.server_thread = threading.Thread(target=self.server.start)
		self.client = Client("localhost", port)

		for i in range(4):
			print()
		print("__________________________________________")

	def tearDown(self):
		self.server.close()
		self.client.end_session()

	def test_practice(self):
		self.server_thread.start()
		# index = "0101101010111001011001000101101011110100001110011101000101111111110001101010111011101101101001100011001101001111000001011010101010001010111000110111010010110010110110101101111010010111101010110001011101000010000100001110011000101110000000010001111010100111"
		# data = "3SL370G8"
		index = util.random_index()
		data = util.random_string()
		self.client.insert_leaf(index, data)
		root = self.client.get_signed_root()
		proof = self.client.generate_proof(index)
		self.assertTrue(self.client.verify_proof(proof, data, root))

		# index = util.random_index()
		# data = util.random_string()
		# for i in range(1):
		# 	index = util.random_index()
		# 	data = util.random_string()
		# 	tx = Client.insert_leaf(index, data)
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
