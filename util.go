package main

func getBaseNameFromImport(path bytes) bytes {
	if strings_Congtains(path, S("/")) {
		words := strings_Split(path, S("/"))
		r := words[len(words)-1]
		return r
	} else {
		return path
	}

}

func getIndex(item bytes, list []bytes) int {
	for id, v := range list {
		if eq(v, item) {
			return id
		}
	}
	return -1
}

func inArray(item bytes, list []bytes) bool {
	for _, v := range list {
		if eq(v, item) {
			return true
		}
	}
	return false
}
