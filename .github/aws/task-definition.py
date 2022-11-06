import json
import os
import sys


class Environment(object):
    DeploymentEnv = None
    ServiceName = None

    AwsDefaultRegion = None
    AwsSecretAccessKey = None
    AwsAccessKeyId = None
    AwsAccountId = None

    ServiceImageRegistry = None
    ServiceImageRepository = None

    ImageTag = None
    DeploymentVersion = None


def create_task_definition():
    env = read_envs()

    secret_arn_base = "arn:aws:secretsmanager:{0}:{1}:secret".format(env.AwsDefaultRegion, env.AwsAccountId)

    credentials_arn_base = "{0}:services/{1}/credentials-{2}:{3}::"

    if env.DeploymentEnv == "production":
        credentials_arn_env = "UUCjIe"
    else:
        credentials_arn_env = ""

    definition_file = {
        "volumes": [],
        "family": "{0}-td-{1}".format(env.ServiceName, env.DeploymentEnv),
        "requiresCompatibilities": [
            "FARGATE"
        ],
        "cpu": "256",
        "memory": "512",
        "executionRoleArn": "arn:aws:iam::{0}:role/ecsTaskExecutionRole".format(env.AwsAccountId),
        "taskRoleArn": "arn:aws:iam::{0}:role/volunteeringServiceTaskRole".format(env.AwsAccountId),
        "networkMode": "awsvpc",
        "containerDefinitions": [
            {
                "essential": True,
                "logConfiguration": {
                    "logDriver": "awslogs",
                    "options": {
                        "awslogs-group": "/ecs/{0}-{1}".format(env.ServiceName, env.DeploymentEnv),
                        "awslogs-region": "{0}".format(env.AwsDefaultRegion),
                        "awslogs-create-group": "true",
                        "awslogs-stream-prefix": "ecs"
                    }
                },
                "portMappings": [
                    {
                        "hostPort": 8002,
                        "protocol": "tcp",
                        "containerPort": 8002
                    }
                ],
                "volumesFrom": [],
                "image": "{0}/{1}:{2}".format(env.ServiceImageRegistry, env.ServiceImageRepository, env.ImageTag),
                "command": ["api"],
                "name": "{0}-{1}".format(env.ServiceName, env.DeploymentEnv),
                "environment": [
                    {
                        "name": "APP_ENV",
                        "value": "{0}".format(env.DeploymentEnv)
                    },
                    {
                        "name": "SERVICE_NAME",
                        "value": "{0}".format(env.ServiceName)
                    },
                    {
                        "name": "AWS_DEFAULT_REGION",
                        "value": "{0}".format(env.AwsDefaultRegion)
                    },
                    {
                        "name": "DEPLOYMENT_VERSION",
                        "value": "{0}".format(env.DeploymentVersion)
                    },
                ],
                "secrets": [
                    {
                        "valueFrom": credentials_arn_base.format(secret_arn_base, env.ServiceName, credentials_arn_env, "PORT"),
                        "name": "PORT"
                    },
                    {
                        "valueFrom": credentials_arn_base.format(secret_arn_base, env.ServiceName, credentials_arn_env, "CLIENT_ID"),
                        "name": "CLIENT_ID"
                    },
                    {
                        "valueFrom": credentials_arn_base.format(secret_arn_base, env.ServiceName, credentials_arn_env, "CLIENT_SECRET"),
                        "name": "CLIENT_SECRET"
                    },
                    {
                        "valueFrom": credentials_arn_base.format(secret_arn_base, env.ServiceName, credentials_arn_env, "SERVICE_ACCOUNT_1"),
                        "name": "SERVICE_ACCOUNT_1"
                    },
                    {
                        "valueFrom": credentials_arn_base.format(secret_arn_base, env.ServiceName, credentials_arn_env, "SERVICE_ACCOUNT_2"),
                        "name": "SERVICE_ACCOUNT_2"
                    },
                ],
            }        
        ]
    }

    with open('.github/aws/task-definition.json', 'w') as f:
        json.dump(definition_file, f)


def read_envs():
    env = Environment()

    env.ServiceName = os.getenv("SERVICE_NAME", None)
    if env.ServiceName is None:
        sys.exit("SERVICE_NAME variable is missing")

    env.DeploymentEnv = os.getenv("DEPLOYMENT_ENV", None)
    if env.DeploymentEnv is None:
        sys.exit("DEPLOYMENT_ENV variable is missing")

    env.ServiceImageRegistry = os.getenv("AWS_IMAGE_REGISTRY", None)
    if env.ServiceImageRegistry is None:
        sys.exit("AWS_IMAGE_REGISTRY variable missing")

    env.ServiceImageRepository = os.getenv("AWS_IMAGE_REPOSITORY", None)
    if env.ServiceImageRepository is None:
        sys.exit("AWS_IMAGE_REPOSITORY variable missing")

    env.AwsAccountId = os.getenv("AWS_ACCOUNT_ID", None)
    if env.AwsAccountId is None:
        sys.exit("AWS_ACCOUNT_ID variable is missing")

    env.AwsDefaultRegion = os.getenv("AWS_DEFAULT_REGION", None)
    if env.AwsDefaultRegion is None:
        sys.exit("AWS_DEFAULT_REGION variable is missing")

    env.AwsSecretAccessKey = os.getenv("AWS_SECRET_ACCESS_KEY", None)
    if env.AwsSecretAccessKey is None:
        sys.exit("AWS_SECRET_ACCESS_KEY variable is missing")

    env.AwsAccessKeyId = os.getenv("AWS_ACCESS_KEY_ID", None)
    if env.AwsAccessKeyId is None:
        sys.exit("AWS_ACCESS_KEY_ID variable is missing")

    env.ImageTag = os.getenv("IMAGE_TAG", None)
    if env.ImageTag is None:
        sys.exit("IMAGE_TAG variable is missing")

    tag_name = os.getenv("SEMAPHORE_GIT_TAG_NAME", None)
    env.DeploymentVersion = env.ImageTag
    if tag_name:
        env.DeploymentVersion = tag_name.lower().split("-")[1] + "+" + env.ImageTag
    if env.DeploymentVersion is None:
        sys.exit("DEPLOYMENT_VERSION variable is missing")

    return env


if __name__ == '__main__':
    create_task_definition()
