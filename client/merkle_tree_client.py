import socket
import json

import server.transaction
from common.util import send_data

class MerkleTreeClient(object):
	def __init__(self, host: str, port: int):
		self.host = host
		self.port = port
		self.socket = socket.socket()
		self.socket.connect((host, port))

	def _listen(self):
		data_length = int(self.socket.recv(200))
		data_type = str(self.socket.recv(200))

		tmp = msg = self.socket.recv(200)
		data_length -= len(tmp)

		while data_length > 0:	
			tmp = self.socket.recv(200)
			msg += tmp
			data_length -= len(tmp)

		return json.loads(msg)

	def practice(self) -> server.transaction.Transaction:
		tx = server.transaction.ReadTransaction("practice")
		send_data(self.socket, tx)
		return server.transaction.Transaction.from_dict(self._listen())

	def get_leaf(self):
		raise NotImplementedError

	def get_proof(self):
		raise NotImplementedError

	def get_signed_root(self):
		raise NotImplementedError

	def verify_proof(self):
		raise NotImplementedError

	def update_leaf(self):
		raise NotImplementedError

	def insert_leaf(self):
		raise NotImplementedError

	def end_session(self):
		print("Ending session...")
		self.socket.close()