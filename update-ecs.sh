DESIRED_COUNT=`aws ecs describe-services --services ${SERVICE_NAME} --cluster ${CLUSTER} --region ${REGION} | jq .services[].desiredCount`;

if [ -z "${DESIRED_COUNT}" ] || [ ${DESIRED_COUNT} = "0" ]; then
	DESIRED_COUNT="1";
fi;

echo "Repository Image URI: ${IMAGE_URI}";
echo "Desired count: ${DESIRED_COUNT}";

aws ecs describe-task-definition --task-definition ${TASK_NAME} | jq '.taskDefinition|{family, containerDefinitions}' -c -M > taskdef.json


echo "Registering the task definition...";
echo $TASK_DEFINITION;

aws ecs register-task-definition --family ${TASK_NAME} --cli-input-json file://`pwd`/taskdef.json --region ${REGION};

REVISION=`aws ecs describe-task-definition --task-definition ${TASK_NAME} --region ${REGION} | jq .taskDefinition.revision`;
echo "New revision is:${REVISION}";

echo "Updating the service (${SERVICE_NAME}) defitinion...";
aws ecs update-service --cluster ${CLUSTER} --region ${REGION} --service ${SERVICE_NAME} --task-definition ${TASK_NAME}:${REVISION} --desired-count ${DESIRED_COUNT};
