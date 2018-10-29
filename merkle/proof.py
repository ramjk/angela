from common import util

class Proof(object):
	def __init__(self, proof_type: bool=True, index: str='', proof_id: util.bitarray=None, copath: list=None) -> None:
		super(Proof, self).__init__()
		
		self.proof_type = proof_type
		self.index = index 
		self.proof_id = proof_id
		self.copath = []
		if copath != None:
			self.copath = copath