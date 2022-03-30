package storage

type Storage map[string]string

func New() Storage {
	return Storage{}
}

func (st Storage) Append(key, value string) {
	st[key] = value
}

/*
type Storage struct {
	File *os.File
	Map  map[string]string
}

func New(holder string) (*Storage, error) {
	file, err := os.OpenFile(holder, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	st := new(Storage)
	st.Map = map[string]string{}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, ";") {
			pair := strings.Split(line, ";")
			st.Map[pair[0]] = pair[1]
		}
	}

	st.File = file
	return st, nil
}

func (st *Storage) Append(key, value string) error {
	st.Map[key] = value
	_, err := st.File.Write([]byte(fmt.Sprintf("%s;%s\n", key, value)))
	if err != nil {
		return err
	}

	return nil
}
*/
