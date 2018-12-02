import socket

class MerkleTreeClient(object):
	def __init__(self, host: str, port: int):
		self.host = host
		self.port = port
		self.socket = socket.socket()
		self.socket.connect((host, port))

	def practice(self) -> str:
		self.socket.send(b"practice")
		return self.socket.recv(16).decode("utf-8")

	def get_leaf(self):
		raise NotImplementedError

	def get_proof(self):
		raise NotImplementedError

	def get_signed_root(self):
		raise NotImplementedError

	def verify_proof(self):
		raise NotImplementedError

	def end_session(self):
		print("Ending session...")
		self.socket.close()