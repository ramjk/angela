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

if __name__ == '__main__':
	unittest.main()		
