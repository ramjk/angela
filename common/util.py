from bitarray import bitarray
import random, string

"""
FIXME: Is this good?
"""
class bitarray(bitarray):
    def __hash__(self):
        return self.tobytes().__hash__()

def to_bytes(data) -> bytes:
	if type(data) == bytes:
		return data
	return data.encode()

def random_index(digest_size: int=256) -> str:
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
