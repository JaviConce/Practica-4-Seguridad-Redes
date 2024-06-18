import requests
import urllib3
import json
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

url="https://myserver.local:5000"
credentials={"username":"TestPython","password":"testpython"}

def test_get_version():
    target=f'{url}/version'
    response=requests.get(target, verify=False)
    print(f"GET {target} - Status Code: {response.status_code}, Response: {response.json()}")
    
def test_post_signup(data):
    target = f'{url}/signup'
    response = requests.post(target, json=data, headers={"Content-Type": "application/json"}, verify=False)
    if response.status_code == 200 and not response.text:
        print(f"POST {target} - Status Code: {response.status_code}, Response: {response.json()}")
        return response.json().get('access_token')
    else:
        print(f"POST {target} - Status Code: {response.status_code}, Response: {response.text}")
        return None
    
    
def test_post_login(data):
    target = f'{url}/login'
    response = requests.post(target, json=data, headers={"Content-Type": "application/json"}, verify=False)
    print(f"POST {target} - Status Code: {response.status_code}, Response: {response.json()}")
    if response.status_code == 200:
        return response.json().get('access_token')
    
def test_post_user_doc_id(data,username,docId):
    target=f'{url}/{username}/{docId}'
    token=test_post_login(credentials)
    headers = {'Authorization':f'{token}', 'Content-Type': 'application/json' }
    response = requests.post(target, data=data,headers=headers, verify=False) 
    if response.text:  # Comprueba si la respuesta no está vacía
        print(f"POST {target} - Status Code: {response.status_code}, Response: {response.text}")
    else:
        print(f"POST {target} - Status Code: {response.status_code}, Response: No response")
    
def test_get_user_doc_id(username,docId):
    target=f'{url}/{username}/{docId}'
    token=test_post_login(credentials)
    headers = {'Authorization':f'{token}', 'Content-Type': 'application/json' }
    response = requests.post(target, headers=headers, verify=False) 
    print(f"GET {target} - Status Code: {response.status_code}, Response: {response.text}")
    
def test_put_user_doc_id(data,username,docId):
    target=f'{url}/{username}/{docId}'
    token=test_post_login(credentials)
    headers = {'Authorization':f'{token}', 'Content-Type': 'application/json' }
    response= requests.put(target,data=data,headers=headers, verify=False)
    print(f"PUT {target} - Status Code: {response.status_code}, Response: {response.text}")
    
def test_delete_user_doc_id(username,docId):
    target=f'{url}/{username}/{docId}'
    token=test_post_login(credentials)
    headers = {'Authorization':f'{token}'}
    response= requests.delete(target, headers=headers, verify=False)
    print(f"DELETE {target} - Status Code: {response.status_code}, Response: {response.text}")
    
def test_get_user_alldocs(username):
    target=f'{url}/{username}/_all_docs'
    token=test_post_login(credentials)
    headers = {'Authorization':f'{token}'}
    response= requests.get(target, headers=headers, verify=False)
    print(f"GET {target} - Status Code: {response.status_code}, Response: {response.text}")
    
if __name__=="__main__":
    obj={"content":"Prueba python Post"}
    json_str=json.dumps(obj)
    test_get_version()
    test_post_signup(credentials)
    test_post_login(credentials)
    test_post_user_doc_id(json_str,"TestPython","prueba")
    test_get_user_doc_id("TestPython","prueba")
    obj={"content":"Prueba python Put"}
    json_str=json.dumps(obj)
    test_put_user_doc_id(json_str,"TestPython","prueba")
    test_get_user_doc_id("TestPython","prueba")
    test_delete_user_doc_id("TestPython","prueba")
    test_get_user_alldocs("TestPython")