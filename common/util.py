import random
from bitarray import bitarray
import random, string

"""
FIXME: Is this good?
"""
class bitarray(bitarray):
    def __hash__(self):
        return self.tobytes().__hash__()

def to_bytes(data: str) -> bytes:
	if type(data) == bytes:
		return data
	return data.encode()

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
