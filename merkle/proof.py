class Proof(object):
	def __init__(self, proof_type, index, copath=None):
		super(Proof, self).__init__()
		
		self.proof_type = proof_type
		self.index = index
		self.copath = []
		if copath != None:
			self.copath = copath