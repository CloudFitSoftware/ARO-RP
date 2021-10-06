set -ex

echo "Creating required directories"

mkdir -p $OB_OUTPUTDIRECTORY/ServiceGroupRoot/bin/
mkdir -p $OB_OUTPUTDIRECTORY/ServiceGroupRoot/Parameters/
mkdir -p $OB_OUTPUTDIRECTORY/Shell/

echo "Downloading Crane"

wget -O $OB_OUTPUTDIRECTORY/Shell/crane.tar.gz https://github.com/google/go-containerregistry/releases/download/v0.4.0/go-containerregistry_Linux_x86_64.tar.gz

echo "Extracting Crane binaries"

pushd $OB_OUTPUTDIRECTORY/Shell
tar xzvf crane.tar.gz
rm crane.tar.gz
popd

echo "Copying required files to ob_outputdirectory: ${OB_OUTPUTDIRECTORY}"

echo "ls -la"
ls -la

echo "ls -lart ./ARO.Pipelines/ev2/generator/"
echo "run du ./ARO.Pipelines/ev2/generator/"
du -h ./ARO.Pipelines/ev2/generator/

echo "run du $OB_OUTPUTDIRECTORY/Shell"
du -h $OB_OUTPUTDIRECTORY/Shell/

which tar
tar --version

du -h ./RP-Config/deploy/ffint-config.yaml
du -h ./RP-Config/deploy/ffprod-config.yaml
# tar -rvf ./ARO.Pipelines/ev2/generator/deployment.tar -C "$OB_OUTPUTDIRECTORY/Shell" $(cd $OB_OUTPUTDIRECTORY/Shell; echo *)
tar -rvf ./ARO.Pipelines/ev2/generator/deployment.tar -C "./RP-Config/deploy" ffint-config.yaml
tar -rvf ./ARO.Pipelines/ev2/generator/deployment2.tar -C "$OB_OUTPUTDIRECTORY/Shell" $(cd $OB_OUTPUTDIRECTORY/Shell; echo *)

tar --concatenate --file=/ARO.Pipelines/ev2/generator/deployment.tar ./ARO.Pipelines/ev2/generator/deployment2.tar

echo "Copy tar to ob_outputdirectory dir"
cp -r ./ARO.Pipelines/ev2/Deployment/ServiceGroupRoot/ $OB_OUTPUTDIRECTORY/
cp ./ARO.Pipelines/ev2/generator/deployment.tar $OB_OUTPUTDIRECTORY/ServiceGroupRoot/bin/

echo "Listing the contents of dirs for debugging"
ls $OB_OUTPUTDIRECTORY
ls $OB_OUTPUTDIRECTORY/ServiceGroupRoot/
ls $OB_OUTPUTDIRECTORY/ServiceGroupRoot/bin/
