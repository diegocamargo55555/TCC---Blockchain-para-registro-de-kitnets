import ipfshttpclient
import psycopg2


def upload_to_ipfs(file_path):
    client = ipfshttpclient.connect('/ip4/127.0.0.1/tcp/5001')
    res = client.add(file_path)
    return res['Hash'] 
    

def registrar_no_banco(proprietario, endereco, cid):
    conn = psycopg2.connect("dbname=a_blockchain user=teste")
    cur = conn.cursor()
    
    query = """
    INSERT INTO kitnets (proprietario_id, endereco_fisico, ipfs_contrato_cid)
    VALUES (%s, %s, %s);
    """
    cur.execute(query, (proprietario, endereco, cid))
    
    conn.commit()
    cur.close()
    conn.close()
    
cid = upload_to_ipfs("img.png")
registrar_no_banco("Kayne","German", cid)