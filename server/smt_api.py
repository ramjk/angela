from ctypes import *
from typing import List
import random
import string
from common import util
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

def batch_insert(keys, values) -> List[bool]:
	raise NotImplementedError

def read(index) -> util.Proof:
	raise NotImplementedError

length = 5
k = [random_index().encode() for i in range(length)]
v = [random_string().encode() for i in range(length)]
k_arr = (c_char_p * len(k))(*k)
v_arr = (c_char_p * len(v))(*v)
keys = GoSlice(k_arr, length, length)
values = GoSlice(v_arr, length, length)
lib.BatchWrite.restype = ndpointer(dtype=c_bool, shape=(length,))
print(lib.BatchWrite(keys, values))
c_char_p_p = POINTER(c_char_p)
lib.argtypes = [c_char_p]
lib.Read.restype = c_char_p_p
result = lib.Read(random_index().encode())
vals = [result[i].decode() for i in range(5)]
print(vals)
lib.FreeCPointers.argtypes = [c_char_p_p, c_int]
lib.FreeCPointers(result, len(vals))
