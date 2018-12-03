from ctypes import *
import random
import string
from bitarray import bitarray

lib = cdll.LoadLibrary("../go/src/main/smt_api.so")
from numpy.ctypeslib import ndpointer

class GoSlice(Structure):
    _fields_ = [("data", POINTER(c_char_p)),
                ("len", c_longlong), ("cap", c_longlong)]

lib.BatchWrite.argtypes = [GoSlice, GoSlice]

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

keys = GoSlice((c_char_p * 5)(c_char_p(random_index().encode()), c_char_p(random_index().encode()), c_char_p(random_index().encode()), c_char_p(random_index().encode()), c_char_p(random_index().encode())), 5, 5)
values = GoSlice((c_char_p * 5)(c_char_p(random_string().encode()), c_char_p(random_string().encode()), c_char_p(random_string().encode()), c_char_p(random_string().encode()), c_char_p(random_string().encode())), 5, 5)
lib.BatchWrite.restype = c_bool

print(lib.BatchWrite(keys, values))



