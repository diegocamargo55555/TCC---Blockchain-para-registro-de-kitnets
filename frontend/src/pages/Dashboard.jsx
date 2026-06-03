import { useState, useEffect } from 'react';

function Dashboard({ onNavigate }) {
  const [kitnet, setKitnet] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Por enquanto, buscamos a KITNET_01 como demonstração.
    // Numa aplicação real, teríamos um endpoint /kitnets que retorna todas.
    fetch('http://localhost:8081/api/kitnets/KITNET_01')
      .then(res => res.json())
      .then(data => {
        if (!data.error) setKitnet(data);
        setLoading(false);
      })
      .catch(err => {
        console.error(err);
        setLoading(false);
      });
  }, []);

  return (
    <div>
      <h1 id="dashboard-heading" tabIndex="-1">Dashboard de Kitnets</h1>
      <p>Gerencie as kitnets tokenizadas na Hyperledger Fabric.</p>

      {loading ? (
        <p>Carregando blocos...</p>
      ) : kitnet ? (
        <div className="card-grid mt-8">
          <div 
            className="glass-panel kitnet-card" 
            style={{ viewTransitionName: 'kitnet-card-1' }}
          >
            <span className="card-badge">Ativo Blockchain</span>
            <h2 style={{ viewTransitionName: 'kitnet-title-1', width: 'fit-content' }}>
              {kitnet.descricao}
            </h2>
            <p><strong>ID:</strong> {kitnet.id}</p>
            <p><strong>Token:</strong> {kitnet.token_id_blockchain}</p>
            <div className="mt-4">
              <strong>Proprietário:</strong> {kitnet.posses[0].entidade_id} ({kitnet.posses[0].percentual}%)
            </div>
            
            <hr style={{ margin: '16px 0', border: 'none', borderTop: '1px solid var(--color-surface-border)' }} />
            
            <h3>Averbações Recentes</h3>
            {kitnet.averbacoes.map(avb => (
              <div key={avb.id} style={{ marginBottom: '8px' }}>
                <a 
                  href={`http://localhost:8080/ipfs/${avb.cid_ipfs}`} 
                  target="_blank" 
                  rel="noreferrer"
                  style={{ color: 'var(--color-primary)', fontWeight: 600, display: 'block' }}
                >
                  📄 {avb.descricao} (Averbação {avb.id})
                </a>
              </div>
            ))}

            <button className="btn btn-outline mt-4" onClick={() => alert('Abrir modal de Averbação!')}>
              + Nova Averbação
            </button>
          </div>
        </div>
      ) : (
        <div className="glass-panel text-center mt-8">
          <h3>Nenhuma kitnet encontrada</h3>
          <button className="btn mt-4" onClick={() => onNavigate('registro')}>Registrar a primeira</button>
        </div>
      )}
    </div>
  );
}

export default Dashboard;
