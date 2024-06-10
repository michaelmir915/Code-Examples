#ifndef SORTEDMPQ_H
#define SORTEDMPQ_H

#include <stdexcept>
#include <list>
#include "MPQ.h"

/*
 * Minimum Priority Queue based on a linked list
 */
template <typename T>
class SortedMPQ : MPQ<T> {
public:
	// Implement the four funtions (insert, is_empty, min, remove_min) from MPQ.h
	// public:
	 // Remove minimum from MPQ and return it
	T remove_min();
	// Get the minimum from MPQ
	T min();
	// Check if MPQ is empty
	bool is_empty();
	// Insert into MPQ
	void insert(const T& data);
	// To hold the elements use std::list
private:
	std::list<T> sortedList;
	// For remove_min() and min() throw exception if the SortedMPQ is empty. Mimir already has a try/catch block so don't use try/catch block here.
};
	template <typename T>
	T SortedMPQ<T>::remove_min() {
		if (this->is_empty()) {
			throw ("Your queue is empty");
		}
		T firstElement = sortedList.front();
		sortedList.pop_front();
		return firstElement;
	}

	template <typename T>
	T SortedMPQ<T>::min() {
		if (this->is_empty()) {
			throw ("Your queue is empty");
		}
		return sortedList.front();
	}
	
	template <typename T>
	bool SortedMPQ<T>::is_empty() {
		return sortedList.empty();
	}

	template <typename T>
	void SortedMPQ<T>::insert(const T& data) {
		if (this->is_empty() || (data > sortedList.back())) {
			sortedList.push_back(data);
		}
		else {
			
			for (auto i = sortedList.begin(); i != sortedList.end(); ++i) {
				if (*i > data) {
					sortedList.insert(i, data);
					break;
				}
				
				
			}
		}
	}
	

#endif