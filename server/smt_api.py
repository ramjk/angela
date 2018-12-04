from ctypes import *
from typing import List
from common.util import Proof
from bitarray import bitarray
import random
import string
import os

lib = cdll.LoadLibrary("go/src/main/smt_api.so")
from numpy.ctypeslib import ndpointer

# class GoSlice(Structure):
#     _fields_ = [("data", POINTER(c_char_p)),
#                 ("len", c_longlong), ("cap", c_longlong)]

lib.BatchWrite.argtypes = [c_char_p, GoSlice, GoSlice, c_char_p]

c_char_p_p = POINTER(c_char_p)
lib.Read.argtypes = [c_char_p]
lib.Read.restype = c_char_p_p

lib.FreeCPointers.argtypes = [c_char_p_p, c_int]

def batch_insert(prefix, keys, values) -> (List[bool], str):
	batch_length = len(keys)
	lib.BatchWrite.restype = ndpointer(dtype=c_bool, shape=(batch_length,))
	k = [s.encode() for s in keys]
	v = [s.encode() for s in values]
	k_arr = (c_char_p * batch_length)(*k)
	v_arr = (c_char_p * batch_length)(*v)
	ids = GoSlice(k_arr, batch_length, batch_length)
	digests = GoSlice(v_arr, batch_length, batch_length)
	root_pointer = cast(create_string_buffer(257), c_char_p)
	writesOk = lib.BatchWrite((c_char_p)(prefix.encode()), ids, digests, root_pointer)
	print(root_pointer)

# If there is nothing in the tree, this will segfault
def read(index) -> Proof:
	proofLength = 255*2 + 3
	result = lib.Read(index)
	vals = [result[i].decode() for i in range(proofLength)]

	proofDict = {}
	proofDict["ProofType"] = vals[0]
	proofDict["QueryID"] = vals[1]
	proofDict["ProofID"] = vals[2]

	coPath = []
	for i in range(3, proofLength, 2):
		coPath += [{"ID": vals[i], "Digest": vals[i+1]}]
	proofDict["CoPath"] = coPath

	lib.FreeCPointers(result, len(vals))
	return Proof.from_dict(proofDict)

batch_insert("00", ["0"], ["hello"])