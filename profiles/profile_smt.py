import cProfile, pstats, merkle.smt
from common import util

keys = [util.random_index() for i in range(10000)]
values = [util.random_string() for i in range(10000)]
proofs = list()
T = merkle.smt.SparseMerkleTree("sha256")

T_batch = merkle.smt.SparseMerkleTree("sha256")

def profile_inserts():
	pr = cProfile.Profile()

	pr.enable()
	for i in range(10000):
		T.insert(keys[i], values[i])
	pr.disable()

	with open("profiles/profile_inserts.txt", 'w') as output:
		p = pstats.Stats(pr, stream=output)
		p.sort_stats('time')
		p.print_stats()

def profile_prover():
	pr = cProfile.Profile()

	pr.enable()
	for i in range(10000):
		proofs.append(T.generate_proof(keys[i]))
	pr.disable()

	with open("profiles/profile_prover.txt", 'w') as output:
		p = pstats.Stats(pr, stream=output)
		p.sort_stats('time')
		p.print_stats()

def profile_verifier():
	pr = cProfile.Profile()

	pr.enable()
	for proof in proofs:
		T.verify_proof(proof)
	pr.disable

	with open("profiles/profile_verifier.txt", 'w') as output:
		p = pstats.Stats(pr, stream=output)
		p.sort_stats('time')
		p.print_stats()	

def profile_batch_inserts():
	pr = cProfile.Profile()

	pr.enable()
	T_batch.batch_insert({k:v for (k,v) in zip(keys, values)})
	pr.disable()

	with open("profiles/profile_batch_inserts.txt", 'w') as output:
		p = pstats.Stats(pr, stream=output)
		p.sort_stats('time')
		p.print_stats()

if __name__ == '__main__':
	profile_inserts()
	profile_prover()
	profile_verifier()
	profile_batch_inserts()
