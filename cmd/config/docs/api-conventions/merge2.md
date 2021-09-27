# Merge (2-way)

  2-way merges fields from a source to a destination, overriding the destination fields
  where they differ.

  ### Merge Rules

  Fields are recursively merged using the following rules:

  - scalars
    - if present only in the dest, it keeps its value
    - if present in the src and is non-null, take the src value -- if `null`, clear it
    - example src: `5`, dest: `3` => result: `5`

  - non-associative lists -- lists without a merge key
    - if present only in the dest, it keeps its value
    - if present in the src and is non-null, take the src value -- if `null`, clear it
    - example src: `[1, 2, 3]`, dest: `[a, b, c]` => result: `[1, 2, 3]`

  - map keys and fields -- paired by the map-key / field-name
    - if present only in the dest, it keeps its value
    - if present only in the src, it is added to the dest
    - if the field is present in both the src and dest, and the src value is
      `null`, the field is removed from the dest
    - if the field is present in both the src and dest, the value is recursively merged
    - example src: `{'key1': 'value1', 'key2': 'value2'}`,
      dest: `{'key2': 'value0', 'key3': 'value3'}`
      => result: `{'key1': 'value1', 'key2': 'value2', 'key3': 'value3'}`

  - associative list elements -- paired by the associative key
    - if present only in the dest, it keeps its value in the list
    - if present only in the src, it is added to the dest list
    - if the field is present in both the src and dest, the value is recursively merged

  ### Associative Keys

  Associative keys are used to identify "same" elements within 2 different lists, and merge them.
  The following fields are recognized as associative keys:

  [`mountPath`, `devicePath`, `ip`, `type`, `topologyKey`, `name`, `containerPort`]

  Any lists where all of the elements contain associative keys will be merged as associative lists.

  ### Example

  > Source

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 3 # scalar
	  template:
	    spec:
	      containers:  # associative list -- (name)
	      - name: nginx
	        image: nginx:1.7
	        command: ['new_run.sh', 'arg1'] # non-associative list
	      - name: sidecar2
	        image: sidecar2:v1

  > Destination

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 1
	  template:
	    spec:
	      containers:
	      - name: nginx
	        image: nginx:1.6
	        command: ['old_run.sh', 'arg0']
	      - name: sidecar1
	        image: sidecar1:v1

  > Result

	apiVersion: apps/v1
	kind: Deployment
	spec:
	  replicas: 3 # scalar
	  template:
	    spec:
	      containers:  # associative list -- (name)
	      - name: nginx
	        image: nginx:1.7
	        command: ['new_run.sh', 'arg1'] # non-associative list
	      - name: sidecar1
	        image: sidecar1:v1
	      - name: sidecar2
	        image: sidecar2:v1
