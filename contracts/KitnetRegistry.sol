// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

/**
 * @title RegistroDeKitnetsERC721
 * @dev Contrato inteligente para registro e gestão de kitnets e contratos de aluguel usando o padrão ERC-721.
 * Baseado na proposta de TCC "Blockchain para cadastro de kitnet".
 */
contract KitnetRegistry is ERC721URIStorage {
    uint256 private _nextTokenId;

    struct Aluguel {
        address locatario;
        uint256 tempoInicio;
        uint256 tempoFim;
        bool estaAtivo;
    }

    // Mapeamento do ID da kitnet para seu aluguel ativo
    mapping(uint256 => Aluguel) public alugueisAtivos;
    // Mapeamento do ID da kitnet para seu status de ocupação
    mapping(uint256 => bool) public estaAlugada;

    // Eventos para transparência e auditoria
    event KitnetRegistrada(uint256 indexed kitnetId, address indexed proprietario, string metadataURI);
    event AluguelIniciado(uint256 indexed kitnetId, address indexed locatario, uint256 tempoFim);
    event AluguelEncerrado(uint256 indexed kitnetId, address indexed locatario);

    constructor() ERC721("KitnetRegistry", "KTNT") {}

    modifier apenasProprietario(uint256 _kitnetId) {
        require(ownerOf(_kitnetId) == msg.sender, "Nao e o dono da kitnet");
        _;
    }

    /**
     * @dev Registra uma nova kitnet no sistema como um NFT.
     * @param _metadataURI O hash do IPFS contendo a representação digital da kitnet.
     */
    function registrarKitnet(string memory _metadataURI) public returns (uint256) {
        _nextTokenId++;
        uint256 novoIdKitnet = _nextTokenId;

        _mint(msg.sender, novoIdKitnet);
        _setTokenURI(novoIdKitnet, _metadataURI);

        emit KitnetRegistrada(novoIdKitnet, msg.sender, _metadataURI);
        return novoIdKitnet;
    }

    /**
     * @dev Sobrescreve o comportamento padrão de transferência para impedir a venda/transferência 
     * enquanto a kitnet estiver alugada.
     * Nota: Utiliza o hook `_update` do OpenZeppelin v5.
     */
    function _update(address para, uint256 tokenId, address auth) internal virtual override returns (address) {
        address de = _ownerOf(tokenId);
        // Só impede a transferência se não for uma criação (mint) ou destruição (burn)
        if (de != address(0) && para != address(0)) {
            require(!estaAlugada[tokenId], "Nao e possivel transferir a propriedade com a kitnet alugada");
        }
        return super._update(para, tokenId, auth);
    }

    /**
     * @dev Registra um contrato de aluguel para uma kitnet.
     * @param _kitnetId O ID da kitnet a ser alugada.
     * @param _locatario O endereço (wallet) do inquilino.
     * @param _duracao Duração do aluguel em segundos.
     */
    function criarAluguel(uint256 _kitnetId, address _locatario, uint256 _duracao) public apenasProprietario(_kitnetId) {
        require(!estaAlugada[_kitnetId], "Kitnet ja esta alugada");
        require(_locatario != address(0), "Endereco de locatario invalido");

        uint256 tempoFim = block.timestamp + _duracao;
        
        alugueisAtivos[_kitnetId] = Aluguel({
            locatario: _locatario,
            tempoInicio: block.timestamp,
            tempoFim: tempoFim,
            estaAtivo: true
        });

        estaAlugada[_kitnetId] = true;

        emit AluguelIniciado(_kitnetId, _locatario, tempoFim);
    }

    /**
     * @dev Encerra um contrato de aluguel ativo.
     * @param _kitnetId O ID da kitnet.
     */
    function encerrarAluguel(uint256 _kitnetId) public {
        address proprietario = ownerOf(_kitnetId); // implicitamente verifica se o token existe
        Aluguel storage aluguel = alugueisAtivos[_kitnetId];
        
        require(aluguel.estaAtivo, "Nao ha aluguel ativo para esta kitnet");
        require(
            msg.sender == proprietario || msg.sender == aluguel.locatario,
            "Apenas o dono ou o locatario podem encerrar o aluguel"
        );

        address locatario = aluguel.locatario;
        aluguel.estaAtivo = false;
        estaAlugada[_kitnetId] = false;

        emit AluguelEncerrado(_kitnetId, locatario);
    }

    /**
     * @dev Recupera informações da kitnet. O proprietário já é nativo do ERC721.
     * @param _kitnetId O ID da kitnet.
     */
    function obterStatusKitnet(uint256 _kitnetId) public view returns (
        address proprietario,
        string memory metadataURI,
        bool alugadaAtualmente
    ) {
        // ownerOf implicitamente verifica se o token existe
        return (ownerOf(_kitnetId), tokenURI(_kitnetId), estaAlugada[_kitnetId]);
    }
}