import unittest
import threading
import ray

from server.transaction import WriteTransaction
from client.merkle_tree_client import MerkleTreeClient
from server.server import Server

class TestSparseMerkleTree(unittest.TestCase):
	def setUp(self):
		port = 3307
		num_worker = 9
		epoch_length = 1000
		tree_depth = 256

		self.server = Server(port, num_worker, epoch_length, tree_depth)
		self.client = MerkleTreeClient("localhost", port)

		for i in range(4):
			print()
		print("__________________________________________")

	def tearDown(self):
		self.server.close()
		self.client.end_session()

	def test_practice(self):
		conn, addr = self.server.socket.accept()
		server_thread = threading.Thread(target=self.server.receive, args=(conn,))
		server_thread.start()
		msg = self.client.practice()
		self.assertEqual(msg, "practice")

	def test_write_transaction(self):
		self.server.epoch_length = 2
		self.server.receive_transaction(WriteTransaction('1000', 'apple'))
		self.server.receive_transaction(WriteTransaction('1001', 'banana'))

if __name__ == '__main__':
	unittest.main()
