
class Transaction(object):
	def __init__(self, transaction_type: str, index: str) -> None:
		self.transaction_type = transaction_type
		self.index = index

	def __eq__(self, other):
		return self.index == other.index

	def __lt__(self, other):
		return self.index < other.index


class ReadTransaction(Transaction):
	def __init__(self, index: str) -> None:
		Transaction.__init__(self, 'R', index)


class WriteTransaction(Transaction):
	def __init__(self, index: str, data: str) -> None:
		Transaction.__init__(self, 'W', index)
		self.data = data
