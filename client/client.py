import socket
import json
import requests

from server import transaction
from common import util

class Client(object):
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

	# def practice(self) -> server.transaction.Transaction:
		# tx = server.transaction.WriteTransaction('1001', "practice")
		# send_data(self.socket, tx)
		# return server.transaction.Transaction.from_dict(self._listen())

	def practice():
		r = requests.get("http://localhost:8000/merkletree")
		return transaction.Transaction.from_dict(json.loads(r.text))

	def get_leaf(self):
		raise NotImplementedError

	def generate_proof(index):
		tx = transaction.ReadTransaction(index)
		r = requests.get("http://localhost:8000/merkletree/proof", params=tx.__dict__)
		if (r.status_code == 400):
			return None
		return util.Proof.from_dict(r.json())

	def get_signed_root(self):
		raise NotImplementedError

	def verify_proof(self):
		raise NotImplementedError

	def insert_leaf(self, index, data):
		transaction = WriteTransaction(index, data)
		util.send_data(self.socket, transaction)
		success = _listen()

	def end_session(self):
		print("Ending session...")
		self.socket.close()
