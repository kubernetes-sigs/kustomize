package merge3_test

var multipleMergeKeysTestCases = []testCase{
	//
	// Test Case
	//
	{
		description: `Update protocol for a port`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`},

	{
		description: `Retain local protocol`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: HTTP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`},

	{
		description: `Append container port`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 80
          protocol: HTTP
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 80
          protocol: HTTP
        - protocol: TCP
          containerPort: 8080
`},

	{
		description: `Update container-port name`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: foo
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
          name: bar
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
          name: foo
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: TCP
          name: bar
`},

	{
		description: `Merge with name for same container-port`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: TCP
          name: updated
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: HTTP
          name: local
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: TCP
          name: updated
        - containerPort: 8080
          name: local
          protocol: HTTP
`},

	{
		description: `Merge with name for same container-port 2`,
		origin: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
`,
		update: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: UDP
          name: updated
`,
		local: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: UDP
          name: local
`,
		expected: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - image: test-image
        name: test-deployment
        ports:
        - containerPort: 8080
          protocol: UDP
          name: original
        - containerPort: 8080
          protocol: UDP
          name: updated
        - containerPort: 8080
          protocol: UDP
          name: local
`},
}
