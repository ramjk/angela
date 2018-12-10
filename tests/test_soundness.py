import subprocess
import time
import os

from math import log
from common import util
from client.client import Client

max_batch = 1024
batches = [2**i for i in range(6, int(log(max_batch, 2)) + 1)]
global_soundness_error = 0
epoch_number = 1

for batch_size in batches:
	soundness_error = 0
	soundness_total = 0
	print("[batch_size {}]".format(batch_size))
	pid = subprocess.Popen(['python', '-m', 'server.flask_server', str(9), str(batch_size), str(256), str(epoch_number)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
	time.sleep(5)
	iterations = int(max_batch / batch_size)
	for i in range(iterations):
		print("    [Iteration {}]".format(i))

		inserts = 0
		indices = list()
		datas = list()

		while inserts < batch_size:
			is_member = util.flip_coin(bias=0.75)
			index = util.random_index()
			data = util.random_string()
			if is_member:
				Client.insert_leaf(index, data)
			inserts += 1
			indices.append(index)
			datas.append(data)
			soundness_total += 1
		root = Client.get_signed_root()
		for i in range(batch_size):
			proof = Client.generate_proof(indices[i])
			data = datas[i]
			if not Client.verify_proof(proof, data, root):
				# print("Proof Type IN TEST SOUNDNESS", proof.ProofType)
				print("FAILED")
				soundness_error += 1

		epoch_number += 1

	pid.kill()
	print("[batch_size {}] {} iterations ==> soundness error {}".format(batch_size, iterations, float(soundness_error) / float(soundness_total)))
	global_soundness_error += soundness_error

print("\nOverall soundness error of {}".format(global_soundness_error))

