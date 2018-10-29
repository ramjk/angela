import math
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

#leaves must be sorted
def find_conflicts(leaves):
	map = {}
	for i in range(1, len(leaves)):
		x, y = leaves[i-1], leaves[i]
		z = x^y
		k = math.floor(math.log2())
		map[x[0:len(x) - k]] = True
	return map
