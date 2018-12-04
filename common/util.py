import random, string
import json
import hashlib
import base64

from bitarray import bitarray

PACKET_SIZE = 200

"""
FIXME: Is this good?
"""
class bitarray(bitarray):
    def __hash__(self):
        return self.tobytes().__hash__()

class Proof(object):
	def __init__(self):
		self.ProofType = None
		self.QueryID = None
		self.ProofID = None
		self.ProofLength = None
		self.CoPath = None

	def from_dict(json_dict):
		proof = Proof()
		proof.__dict__ = json_dict
		return proof

def empty(depth):
	if depth == 0:
		return SHA256("")
	t = empty(depth - 1) 
	return SHA256(t + t)	

def to_bytes(data: str) -> bytes:
	if type(data) == bytes:
		return data
	return base64.b64decode(data)

def to_string(data: bytes) -> str:
	if type(data) == str:
		return str
	return base64.b64encode(data).decode()

def SHA256(data):
	data = to_bytes(data)
	h = hashlib.sha256()
	h.update(data)
	return h.digest()

#leaves must be sorted
def find_conflicts(leaves: list):
	conflicts = {}
	for i in range(1, len(leaves)):
		x, y = bitarray(leaves[i-1]), bitarray(leaves[i])
		z = x^y
		k = len(x)
		for idx, elem in enumerate(z):
			if elem:
				k = idx
				break
		conflicts[x[0:k]] = True
	return conflicts

def random_index(digest_size: int = 256) -> str:
	bitarr = list()
	for i in range(digest_size):
		r = random.random()
		if r > 0.5:
			bitarr.append(1)
		else:
			bitarr.append(0)
	bitstring = bitarray(bitarr).to01()
	return bitstring

def random_string(size: int=8) -> str:
	return ''.join(random.choices(string.ascii_uppercase + string.digits, k=size))

def flip_coin():
	r = random.random()
	if r > 0.5:
		return True
	return False

def left_sibling(index):
	return index[-1] == "0"

def pad_packet(packet):
	return packet.ljust(PACKET_SIZE)

def send_data(sock, data: object):
	json_dump = json.dumps(data.__dict__)
	sock.sendall(pad_packet(str(len(json_dump)).encode()))
	sock.sendall(pad_packet(b"data-type"))
	sock.sendall(pad_packet(json_dump.encode()))
