from bitarray import bitarray

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
