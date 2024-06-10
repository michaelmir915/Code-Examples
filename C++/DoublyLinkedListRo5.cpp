// implementation of the DLList class
#include "DLList.h"
//Node is already declared in header

//Linked lists
//Rule of five:
// 
//Constructor
DLList::DLList() {
	header.next = &trailer;
	trailer.prev = &header;
}

//Copy constructor
DLList::DLList(const DLList& dll) {
	//set current list to be empty insert last a bunch of times
	this->header.next = &this->trailer;
	this->trailer.prev = &this->header; 
	DLListNode* tempNode = dll.header.next;
	if (tempNode != &dll.trailer) {
		while (tempNode != &dll.trailer) {
			this->insert_last(tempNode->obj);
			tempNode = tempNode->next;
		}
	}
}

// move constructor
DLList::DLList(DLList&& dll) {
	this->header.next = dll.header.next;
	this->header.next->prev = &this->header;
	this->trailer.prev = dll.trailer.prev;
	this->trailer.prev->next = &this->trailer;
	dll.header.next = &dll.trailer;
	dll.trailer.prev = &dll.header;
	
}

//Destructor
DLList::~DLList() {
	DLListNode* prevNode, * tempNode = header.next;
	while (tempNode != &trailer) {
		prevNode = tempNode;
		tempNode = tempNode->next;
		delete prevNode; //ASK: why dont we delete tempNode?
	}
	header.next = &trailer;
	trailer.prev = &header;
}

//Copy assignemnt operator
DLList& DLList::operator=(const DLList& dll) {
	//self assign check 
	if (this != &dll) {
		this->make_empty();
		this->header.next = &this->trailer;
		this->trailer.prev = &this->header; 
		DLListNode* tempNode = dll.header.next;
		if (tempNode != &dll.trailer) {
			while (tempNode != &dll.trailer) {
				this->insert_last(tempNode->obj);
				tempNode = tempNode->next;
			}
		}
	}
	return *this;
}

//Move assignment operator
DLList& DLList::operator=(DLList&& dll) {
	//self assign check
	if (this != &dll) {
		this->make_empty();
		this->header.next = dll.header.next;
		this->header.next->prev = &this->header;
		this->trailer.prev = dll.trailer.prev;
		this->trailer.prev->next = &this->trailer;
		dll.header.next = &dll.trailer;
		dll.trailer.prev = &dll.header;
	}
	return *this;
}

// return the pointer to the first node (header's next)
DLList::DLListNode* DLList::first_node() const {
	return header.next;
}

// return the pointer to the trailer
const DLList::DLListNode* DLList::after_last_node() const {
	return &trailer;
}

// return if the list is empty
bool DLList::is_empty() const {
	return (this->header.next == &trailer);
}

// return the first object
int DLList::first() const {
	if (is_empty()) {
		throw ("Your linked list is empty");
	}
	return header.next->obj;
}

// return the last object
int DLList::last() const {
	if (is_empty()) {
		throw ("Your linked list is empty");
	}
	return trailer.prev->obj;
}

// insert to the first node
void DLList::insert_first(const int obj) {
	DLListNode* firstNode = new DLListNode(obj, &header, header.next);
	header.next->prev = firstNode;
	header.next = firstNode;
}

// remove the first node
int DLList::remove_first() {
	if (is_empty()) {
		throw ("Your linked list is empty");
	}
	DLListNode* tempNode = header.next;
	tempNode->next->prev = &header;
	header.next = tempNode->next;
	int returnObj = tempNode->obj;
	delete tempNode;
	return returnObj;
}

// insert to the last node
void DLList::insert_last(const int obj) {
	DLListNode* lastNode = new DLListNode(obj, trailer.prev, &trailer);
	trailer.prev->next = lastNode;
	trailer.prev = lastNode;
}


// remove the last node
int DLList::remove_last() {
    //throw an error if list is empty 
	if (is_empty()) {
		throw ("Your linked list is empty");
	}
	DLListNode* tempNode = trailer.prev;
	tempNode->prev->next = &trailer;
	trailer.prev = tempNode->prev;
	int returnObj = tempNode->obj;
	delete tempNode;
	return returnObj;
}


void DLList::insert_after(DLListNode& p, const int obj) {
	//throw an error if list is empty 
	if (header.next == &trailer) {
		throw ("Specified object does not exist.");
	}
	DLListNode* newNode = new DLListNode(obj, &p, p.next);
	p.next->prev = newNode;
	p.next = newNode;
}
void DLList::insert_before(DLListNode& p, const int obj) {
	//throw an error if list is empty 
	if (header.next == &trailer) {
		throw ("Specified object does not exist.");
	}
	DLListNode* newNode = new DLListNode(obj, p.prev, &p);
	p.prev->next = newNode;
	p.prev = newNode;
}


int DLList::remove_after(DLListNode& p) {
	//throw an error if list is empty or theres nothing after p
	if (header.next == &trailer || trailer.prev == &p) {
		throw ("Specified object does not exist.");
	}
	DLListNode* tempNode = header.next;
	while (tempNode->next != &trailer) {
		if (tempNode == &p) {
			tempNode = tempNode->next;
			tempNode->prev->next = tempNode->next;
			tempNode->next->prev = tempNode->prev;
			int _obj = tempNode->obj;
			delete tempNode;
			return _obj;
		}
		tempNode = tempNode->next;
	}
	throw ("Specified object does not exist.");
}

int DLList::remove_before(DLListNode& p) {
	//throw an error if list is empty or theres nothing before p
	if (header.next == &trailer || header.next == &p) {
		throw ("Specified object does not exist.");
	}
	DLListNode* tempNode = header.next;
	while (tempNode->next != &trailer) {
		if (tempNode->next == &p) {
			tempNode->prev->next = tempNode->next;
			tempNode->next->prev = tempNode->prev;
			int _obj = tempNode->obj;
			delete tempNode;
			return _obj;
		}
		tempNode = tempNode->next;
	}
	throw ("Specified object does not exist.");
}

void DLList::make_empty(void) {
    //Check if list is already empty
	if (header.next == &trailer) {
		return;
	}
	DLListNode* tempNode = header.next;
	do {
		header.next = tempNode->next;
		delete tempNode;
		tempNode = header.next;
	} while (header.next != &trailer);
	trailer.prev = &header;
}

std::ostream& operator<<(std::ostream& out, const DLList& dll) {
	DLList::DLListNode* tempNode = dll.first_node();
	while (tempNode != dll.after_last_node()) {
		out << tempNode->obj << ", ";
		tempNode = tempNode->next;
	}
	return out;
}