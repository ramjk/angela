import socket
import json
import requests
import hashlib
import base64

from server import transaction
from common import util

class Client(object):
	def __init__(self, host: str, port: int):
		self.host = host
		self.port = port
		self.socket = socket.socket()

	def _listen(self):
		pckt = self.socket.recv(200)
		data_length = int(pckt)
		data_type = str(self.socket.recv(200))

		tmp = msg = self.socket.recv(200)
		data_length -= len(tmp)

		while data_length > 0:	
			tmp = self.socket.recv(200)
			msg += tmp
			data_length -= len(tmp)

		print("Received data")
		self.socket.close()
		self.socket = socket.socket()
		return msg

	def practice():
		r = requests.get("http://localhost:8000/merkletree")
		return transaction.Transaction.from_dict(json.loads(r.text))

	def get_leaf(self):
		raise NotImplementedError

	def insert_leaf(index, data):
		# util.connect_and_send(self.socket, self.host, self.port, tx)
		# msg = self._listen()
		# if msg == "True":
		# 	return True
		# return False

		tx = transaction.WriteTransaction(index, data)
		r = requests.post("http://localhost:5000/merkletree/update", data=tx.__dict__)
		print(r.text)

	def generate_proof(index):
		tx = transaction.ReadTransaction(index)
		r = requests.get("http://localhost:5000/merkletree/prove", params=tx.__dict__)
		if (r.status_code == 400):
			return None
		return util.Proof.from_dict(r.json())
		# tx = transaction.ReadTransaction(index)
		# util.connect_and_send(self.socket, self.host, self.port, tx)
		# json_dict = json.loads(self._listen())
		# return util.Proof.from_dict(json_dict)

	def get_signed_root():
		r = requests.get("http://localhost:5000/merkletree/root")
		return r.text

	def verify_proof(proof, data, root):
		proof_id_length = len(proof.ProofID)
		tmp =  None
		if proof.ProofType == False:
			if proof_id_length > len(proof.QueryID):
				return False
			for i in range(proof_id_length):
				if proof.ProofID[i] != proof.QueryID[i]:
					return False
			tmp = util.empty(256 - proof_id_length)
		else:
			tmp = util.SHA256(util.to_bytes(data))

		for node in proof.CoPath:
			if util.left_sibling(node["ID"]):
				tmp = util.SHA256(util.to_bytes(node["Digest"]) + tmp)
			else:
				tmp = util.SHA256(tmp + util.to_bytes(node["Digest"]))
		actual_digest = util.to_string(tmp)
		print("verification", actual_digest)
		print("root", root)
		return actual_digest == root

	def end_session(self):
		print("Ending session...")
		self.socket.close()

if __name__ == '__main__':
	indices = [util.random_index() for i in range(1000)]
	data = [util.random_string() for i in range(1000)]
	
	for i in range(1000):
		print("Insert {}".format(i))
		Client.insert_leaf(indices[i], data[i])

	root = Client.get_signed_root()

	for i in range(1000):
		print("Generate proof {}".format(i))
		proof = Client.generate_proof(indices[i])
		print("Verify proof {}".format(i))
		accept = Client.verify_proof(proof, data[i], root)
		if accept:
			print("Pass!")
		else:
			print("Fail...")
			break
