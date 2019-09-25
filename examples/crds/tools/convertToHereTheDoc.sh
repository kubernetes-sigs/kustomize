#!/bin/bash
DEMONUMBER=openshift-api
DEMONAME=${DEMONUMBER}
NEWFILE=README.md
APPDIR=recreate${DEMONUMBER}
ALLINONEDIR=$HOME/src/sigs.k8s.io/kustomize/examples/crds/${DEMONAME}
DEMODIR=$ALLINONEDIR
KUSTFILES=`find $DEMODIR -name kustomization.yaml -print | sort`
CONFIGFILES=`find $DEMODIR -name kustomizeconfig.yaml -print | sort`
RESOURCES=`find $DEMODIR -name "*.yaml" -print | grep -v expected | grep -v kustomization.yaml | grep -v kustomizeconfig.yaml | sort`
OTHERFILES=`find $DEMODIR -type f -print | grep -v expected | grep -v README.md | grep -v yaml | sort`
EXPECTED=`find $DEMODIR -name "*.yaml" -print | grep expected | sort`
DIRLIST=`find $DEMODIR -type d -print | grep -v expected | sort`

echo "# Test CRD Register "${DEMONUMBER} > $NEWFILE
echo "" >> $NEWFILE

sed -e "s;DEMONUMBER;$DEMONUMBER;g" templates/workspace.txt >> $NEWFILE

echo "## Preparation" >> $NEWFILE
echo "" >> $NEWFILE

echo '<!-- @makeDirectories @test -->' >> $NEWFILE
echo '```bash' >> $NEWFILE
for i in $DIRLIST
do
DIRNAME=`echo $i | sed -e "s;${DEMODIR}/;;g"`
echo 'mkdir -p ${DEMO_HOME}/'${DIRNAME} >> $NEWFILE
done
echo '```' >> $NEWFILE

KUSTCOUNTER=0
for i in $KUSTFILES
do
STEP="KustomizationFile"$((KUSTCOUNTER))
FILENAME=`echo $i | sed -e "s;${DEMODIR}/;;g"`
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/header.txt >> $NEWFILE
cat $i >> $NEWFILE
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/footer.txt >> $NEWFILE
KUSTCOUNTER=$((KUSTCOUNTER +1))
done

CONFIGCOUNTER=0
for i in $CONFIGFILES
do
STEP="KustomizeConfig"$((CONFIGCOUNTER))
FILENAME=`echo $i | sed -e "s;${DEMODIR}/;;g"`
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/header.txt >> $NEWFILE
cat $i >> $NEWFILE
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/footer.txt >> $NEWFILE
CONFIGCOUNTER=$((CONFIGCOUNTER +1))
done

RESOURCECOUNTER=0
for i in $RESOURCES
do
STEP="Resource"$((RESOURCECOUNTER))
FILENAME=`echo $i | sed -e "s;${DEMODIR}/;;g"`
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/header.txt >> $NEWFILE
cat $i >> $NEWFILE
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/footer.txt >> $NEWFILE
RESOURCECOUNTER=$((RESOURCECOUNTER +1))
done

OTHERCOUNTER=0
for i in $OTHERFILES
do
STEP="Other"$((OTHERCOUNTER))
FILENAME=`echo $i | sed -e "s;${DEMODIR}/;;g"`
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/header.txt >> $NEWFILE
cat $i >> $NEWFILE
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/footer.txt >> $NEWFILE
OTHERCOUNTER=$((OTHERCOUNTER +1))
done

echo "## Execution" >> $NEWFILE

cat templates/execution.txt >> $NEWFILE

echo "## Verification" >> $NEWFILE

cat templates/verificationheader.txt >> $NEWFILE

EXPECTEDCOUNTER=0
for i in $EXPECTED
do
STEP="Expected"$((EXPECTEDCOUNTER))
FILENAME=`echo $i | sed -e "s;${DEMODIR}/;;g"`
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" -e "s;Preparation;Verification;g" templates/header.txt >> $NEWFILE
cat $i >> $NEWFILE
sed -e "s;FOOBAR;$FILENAME;g" -e "s;STEP;$STEP;g" templates/footer.txt >> $NEWFILE
EXPECTEDCOUNTER=$((EXPECTEDCOUNTER +1))
done

cat templates/verificationfooter.txt >> $NEWFILE
