"""Create a small AWS estate in a moto server for the Unclick end-to-end test.

Usage: AWS_ENDPOINT_URL=http://127.0.0.1:5000 python seed.py
"""

import os

import boto3

REGION = os.environ.get("AWS_REGION", "us-east-1")


def main() -> None:
    ec2 = boto3.client("ec2", region_name=REGION)
    vpc = ec2.create_vpc(CidrBlock="10.42.0.0/16")["Vpc"]["VpcId"]
    ec2.create_tags(Resources=[vpc], Tags=[{"Key": "Name", "Value": "unclick-e2e"}])
    subnet = ec2.create_subnet(VpcId=vpc, CidrBlock="10.42.1.0/24", AvailabilityZone=REGION + "a")["Subnet"]["SubnetId"]
    ec2.create_tags(Resources=[subnet], Tags=[{"Key": "Name", "Value": "unclick-e2e-a"}])
    sg = ec2.create_security_group(GroupName="unclick-e2e-web", Description="web", VpcId=vpc)["GroupId"]
    ec2.authorize_security_group_ingress(
        GroupId=sg,
        IpPermissions=[{
            "IpProtocol": "tcp",
            "FromPort": 443,
            "ToPort": 443,
            "IpRanges": [{"CidrIp": "0.0.0.0/0", "Description": "https"}],
        }],
    )

    print(f"vpc={vpc} subnet={subnet} sg={sg}")


if __name__ == "__main__":
    main()
