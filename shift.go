package swiss

// Shift increments the value associated with the given key by the specified amount.
// If the key doesn't exist, it inserts a new key-value pair with the amount as the value.
func (m *Map[K, V]) Shift(key K, amount V) {
	value, ok := m.Get(key)
	if !ok {
		// If the key doesn't exist, insert a new key-value pair
		m.Put(key, amount)
		return
	}

	// Perform the addition
	newValue, ok := add(value, amount)
	if !ok {
		// Handle the case where addition is not possible
		panic("Cannot perform addition on the value type")
	}

	m.Put(key, newValue)
}

// Shift2 increments the value associated with the given key by the specified amount.
// If the key doesn't exist, it inserts a new key-value pair with the amount as the value.
// Concept is same as Shift, but Get and Push are now inlined
func (m *Map[K, V]) Shift2(key K, amount V) {
	hi, lo := splitHash(m.hash.Hash(key))
	g := probeStart(hi, len(m.groups))

	// Inline Get operation
	var value V
	var ok bool
	for {
		matches := metaMatchH2(&m.ctrl[g], lo)
		for matches != 0 {
			s := nextMatch(&matches)
			if key == m.groups[g].keys[s] {
				value, ok = m.groups[g].values[s], true
				goto found
			}
		}
		matches = metaMatchEmpty(&m.ctrl[g])
		if matches != 0 {
			ok = false
			goto found
		}
		g++
		if g >= uint32(len(m.groups)) {
			g = 0
		}
	}

found:
	if !ok {
		// If the key doesn't exist, insert the amount as the initial value
		goto insert
	}

	// Perform the addition
	value, ok = add(value, amount)
	if !ok {
		// Handle the case where addition is not possible
		panic("Cannot perform addition on the value type")
	}

	// Inline Put operation
insert:
	if m.resident >= m.limit {
		m.rehash(m.nextSize())
		// Recalculate hash and probe start after rehash
		hi, lo = splitHash(m.hash.Hash(key))
		g = probeStart(hi, len(m.groups))
	}

	for {
		matches := metaMatchH2(&m.ctrl[g], lo)
		for matches != 0 {
			s := nextMatch(&matches)
			if key == m.groups[g].keys[s] {
				m.groups[g].keys[s] = key
				m.groups[g].values[s] = value
				return
			}
		}
		matches = metaMatchEmpty(&m.ctrl[g])
		if matches != 0 {
			s := nextMatch(&matches)
			m.groups[g].keys[s] = key
			m.groups[g].values[s] = value
			m.ctrl[g][s] = int8(lo)
			m.resident++
			return
		}
		g++
		if g >= uint32(len(m.groups)) {
			g = 0
		}
	}
}

// add is a helper function to perform addition on values of type V
func add[V any](a, b V) (V, bool) {
	switch v := any(a).(type) {
	case int:
		if bInt, ok := any(b).(int); ok {
			return any(v + bInt).(V), true
		}
	case int64:
		if bInt64, ok := any(b).(int64); ok {
			return any(v + bInt64).(V), true
		}
	case float64:
		if bFloat64, ok := any(b).(float64); ok {
			return any(v + bFloat64).(V), true
		}
		// Add more cases for other numeric types as needed
	}
	return a, false // Return original value and false if addition not possible
}
