import { useState } from 'react';

function Registro({ onNavigate }) {
  const [loading, setLoading] = useState(false);
  const [kitnetData, setKitnetData] = useState({
    kitnet_id: '',
    kitnet_descricao: '',
    kitnet_token_id: ''
  });
  const [donos, setDonos] = useState([
    { id: '', tipo: 'PF', documento: '', nome: '', carteira: '', representante_legal_id: '', percentual_posse: 100 }
  ]);

  const handleKitnetChange = (e) => {
    setKitnetData({ ...kitnetData, [e.target.name]: e.target.value });
  };

  const handleDonoChange = (index, field, value) => {
    const newDonos = [...donos];
    newDonos[index][field] = value;
    setDonos(newDonos);
  };

  const addDono = () => {
    setDonos([...donos, { id: '', tipo: 'PF', documento: '', nome: '', carteira: '', representante_legal_id: '', percentual_posse: 0 }]);
  };

  const removeDono = (index) => {
    const newDonos = [...donos];
    newDonos.splice(index, 1);
    setDonos(newDonos);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    const totalPosse = donos.reduce((acc, dono) => acc + parseFloat(dono.percentual_posse || 0), 0);
    if (totalPosse !== 100) {
      alert(`A soma dos percentuais de posse deve ser exatamente 100%. Total atual: ${totalPosse}%`);
      return;
    }

    setLoading(true);

    const payload = {
      ...kitnetData,
      donos: donos.map(d => ({
        ...d,
        percentual_posse: parseFloat(d.percentual_posse)
      }))
    };

    try {
      const res = await fetch('http://localhost:8081/api/kitnets', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      
      const result = await res.json();
      if (!res.ok) throw new Error(result.error);
      
      alert(result.message);
      onNavigate('dashboard');
    } catch (err) {
      alert('Erro: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: '700px', margin: '0 auto', paddingBottom: '40px' }}>
      <h1 id="registro-heading" tabIndex="-1">Registrar Imóvel</h1>
      <p>Cadastre uma nova Kitnet no Cartório Chain.</p>

      <form onSubmit={handleSubmit} className="glass-panel mt-8">
        <h2>Dados do Imóvel</h2>
        <div className="form-group">
          <label>ID da Kitnet (ex: KITNET_02)</label>
          <input type="text" name="kitnet_id" className="input-text" value={kitnetData.kitnet_id} onChange={handleKitnetChange} required />
        </div>
        <div className="form-group">
          <label>Descrição</label>
          <input type="text" name="kitnet_descricao" className="input-text" value={kitnetData.kitnet_descricao} onChange={handleKitnetChange} required />
        </div>
        <div className="form-group">
          <label>Token ID</label>
          <input type="text" name="kitnet_token_id" className="input-text" value={kitnetData.kitnet_token_id} onChange={handleKitnetChange} required />
        </div>

        <h2 className="mt-8">Proprietários</h2>
        {donos.map((dono, index) => (
          <div key={index} className="dono-section" style={{ padding: '15px', border: '1px solid #ccc', borderRadius: '8px', marginBottom: '15px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <h3>Dono {index + 1}</h3>
              {donos.length > 1 && (
                <button type="button" onClick={() => removeDono(index)} style={{ background: '#ff4d4f', color: '#fff', border: 'none', padding: '5px 10px', borderRadius: '4px', cursor: 'pointer' }}>
                  Remover
                </button>
              )}
            </div>

            <div className="form-group">
              <label>Tipo de Entidade</label>
              <select className="input-text" value={dono.tipo} onChange={(e) => handleDonoChange(index, 'tipo', e.target.value)} required>
                <option value="PF">Pessoa Física (PF)</option>
                <option value="PJ">Pessoa Jurídica (PJ)</option>
              </select>
            </div>
            <div className="form-group">
              <label>ID da Entidade (ex: ENT002)</label>
              <input type="text" className="input-text" value={dono.id} onChange={(e) => handleDonoChange(index, 'id', e.target.value)} required />
            </div>
            <div className="form-group">
              <label>Nome Completo / Razão Social</label>
              <input type="text" className="input-text" value={dono.nome} onChange={(e) => handleDonoChange(index, 'nome', e.target.value)} required />
            </div>
            <div className="form-group">
              <label>{dono.tipo === 'PF' ? 'CPF' : 'CNPJ'}</label>
              <input type="text" className="input-text" value={dono.documento} onChange={(e) => handleDonoChange(index, 'documento', e.target.value)} required />
            </div>
            <div className="form-group">
              <label>Carteira Blockchain</label>
              <input type="text" className="input-text" value={dono.carteira} onChange={(e) => handleDonoChange(index, 'carteira', e.target.value)} required />
            </div>

            {dono.tipo === 'PJ' && (
              <div className="form-group">
                <label>ID do Representante Legal (obrigatório para PJ)</label>
                <input type="text" className="input-text" value={dono.representante_legal_id} onChange={(e) => handleDonoChange(index, 'representante_legal_id', e.target.value)} required={dono.tipo === 'PJ'} />
              </div>
            )}

            <div className="form-group">
              <label>Percentual de Posse (%)</label>
              <input type="number" step="0.01" className="input-text" value={dono.percentual_posse} onChange={(e) => handleDonoChange(index, 'percentual_posse', e.target.value)} min="0.01" max="100" required />
            </div>
          </div>
        ))}

        <button type="button" onClick={addDono} style={{ background: '#0b57d0', color: '#fff', border: 'none', padding: '10px 15px', borderRadius: '4px', cursor: 'pointer', marginBottom: '20px', width: '100%' }}>
          + Adicionar Proprietário
        </button>

        <button type="submit" className="btn mt-4" disabled={loading} style={{ width: '100%' }}>
          {loading ? 'Gravando no Ledger...' : 'Gravar na Blockchain'}
        </button>
      </form>
    </div>
  );
}

export default Registro;
