import unittest
from common import util

class TestUtil(unittest.TestCase):
	def test_to_bytes(self):
		string = "this is a string"
		bytestring = util.to_bytes(string)
		self.assertEqual(bytes, type(bytestring))

		bytestring = b"this is a bytestring"
		bytestring = util.to_bytes(bytestring)
		self.assertEqual(bytes, type(bytestring))

	def test_find_conflicts(self):
		leaves = [
			bitarray([0, 0, 0, 0]), 
			bitarray([0, 0, 0, 1]), 
			bitarray([0, 0, 1, 0]), 
			bitarray([0, 1, 1, 0]),
			bitarray([1, 1, 1, 0])
				]

		correct_conflict = {
						bitarray(): True, 
						bitarray([0, 0, 0]): True, 
						bitarray([0, 0]): True,
						bitarray([0]): True
							}
		conflicts = util.find_conflicts(leaves)

		self.assertEqual(len(conflicts), len(correct_conflict))
		self.assertDictEqual(conflicts, correct_conflict)


if __name__ == '__main__':
	unittest.main()
