from ctypes import *
from typing import List
from common.util import Proof
import random
import string
import os

lib = cdll.LoadLibrary("/home/ubuntu/angela/go/src/main/smt_api.so")
from numpy.ctypeslib import ndpointer

class GoSlice(Structure):
    _fields_ = [("data", POINTER(c_char_p)),
                ("len", c_longlong), ("cap", c_longlong)]

lib.BatchWrite.argtypes = [c_char_p, GoSlice, GoSlice, c_int]	
lib.BatchWrite.restype = c_char_p

c_char_p_p = POINTER(c_char_p)
lib.Read.argtypes = [c_char_p]
lib.Read.restype = c_char_p_p

lib.GetLatestRoot.restype = c_char_p

lib.FreeCPointers.argtypes = [c_char_p_p, c_int]

def batch_insert(prefix, keys, values, epoch_number) -> str:
	batch_length = len(keys)
	# lib.BatchWrite.restype = ndpointer(dtype=c_bool, shape=(batch_length,))
	k = [s.encode() for s in keys]
	v = [s.encode() for s in values]
	k_arr = (c_char_p * batch_length)(*k)
	v_arr = (c_char_p * batch_length)(*v)
	ids = GoSlice(k_arr, batch_length, batch_length)
	digests = GoSlice(v_arr, batch_length, batch_length)
	root = lib.BatchWrite((c_char_p)(prefix.encode()), ids, digests, epoch_number)
	result = root.decode()
	return result

# If there is nothing in the tree, this will segfault
def read(index) -> Proof:
	numMetaData = 4
	result = lib.Read(index.encode())
	proofDict = {}
	proofDict["ProofType"] = (result[0].decode() == "true")
	# print("ProofType in SMT_API", proofDict["ProofType"])
	proofDict["QueryID"] = result[1].decode()
	proofDict["ProofID"] = result[2].decode()
	proofDict["ProofLength"] = int(result[3].decode())
	# print("correct number should be", proofDict["ProofLength"])
	coPath = []
	for i in range(numMetaData, proofDict["ProofLength"], 2):
		coPath += [{"ID": result[i].decode(), "Digest": result[i+1].decode()}]
	proofDict["CoPath"] = coPath

	lib.FreeCPointers(result, proofDict["ProofLength"])
	return Proof.from_dict(proofDict)

def getLatestRootDigest() -> string:
	root = lib.GetLatestRoot()
	result = root.decode()
	return result
