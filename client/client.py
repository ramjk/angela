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

	def practice():
		r = requests.get("http://localhost:8000/merkletree")
		return transaction.Transaction.from_dict(json.loads(r.text))

	def get_leaf(self):
		raise NotImplementedError

	def generate_proof(index):
		tx = transaction.ReadTransaction(index)
		r = requests.get("http://localhost:8000/merkletree/prove", params=tx.__dict__)
		if (r.status_code == 400):
			return None
		return util.Proof.from_dict(r.json())

	def get_signed_root():
		r = requests.get("http://localhost:8000/merkletree/root")
		return r.json()

	def verify_proof(proof, data, root):
		proof_id_length = len(proof.ProofID)
		print("proof", proof.__dict__)
		print("proofID", proof.ProofID)
		tmp =  None
		if proof.ProofType == False:
			if proof_id_length > len(proof.QueryID):
				return False
			for i in range(proof_id_length):
				if proof.ProofID[i] != proof.QueryID[i]:
					return False
			tmp = util.empty(256 - proof_id_length)
			print("tmp", util.to_string(tmp))
		else:
			tmp = util.SHA256(util.to_bytes(data))

		for node in proof.CoPath:
			if util.left_sibling(node["ID"]):
				tmp = util.SHA256(util.to_bytes(node["Digest"]) + tmp)
			else:
				tmp = util.SHA256(tmp + util.to_bytes(node["Digest"])) 
		actual_digest = util.to_string(tmp)
		print("actual_digest", actual_digest)
		return actual_digest == root

	def insert_leaf(index, data):
		tx = transaction.WriteTransaction(index, data)
		r = requests.post("http://localhost:8000/merkletree/update", data=json.dumps(tx.__dict__))
		return r.status_code

	def end_session(self):
		print("Ending session...")
		self.socket.close()
