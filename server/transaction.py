
class Transaction(object):
	def __init__(self, transaction_type: str, index: str) -> None:
		self.transaction_type = transaction_type
		self.index = index


class ReadTransaction(Transaction):
	def __init__(self, transaction_type: str, index: str) -> None:
		super(transaction_type, index)


class WriteTransaction(Transaction):
	def __init__(self, transaction_type: str, index: str, data: str) -> None:
		super(transaction_type, index)
		self.data = data
