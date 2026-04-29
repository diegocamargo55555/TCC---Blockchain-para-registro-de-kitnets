import { expect } from "chai";
import hre from "hardhat";

describe("KitnetRegistry (Registro de Kitnets)", function () {
  let kitnetRegistry;
  let dono;
  let locatario;
  let outraConta;

  beforeEach(async function () {
    [dono, locatario, outraConta] = await hre.ethers.getSigners();
    const KitnetRegistry = await hre.ethers.getContractFactory("KitnetRegistry");
    kitnetRegistry = await KitnetRegistry.deploy();
  });

  describe("Registro de Kitnet", function () {
    it("Deve registrar uma nova kitnet e emitir o evento KitnetRegistrada", async function () {
      const metadataURI = "ipfs://QmTest123";
      
      await expect(kitnetRegistry.registrarKitnet(metadataURI))
        .to.emit(kitnetRegistry, "KitnetRegistrada")
        .withArgs(1, dono.address, metadataURI);

      expect(await kitnetRegistry.ownerOf(1)).to.equal(dono.address);
      expect(await kitnetRegistry.tokenURI(1)).to.equal(metadataURI);
    });

    it("Deve incrementar o ID para cada novo registro", async function () {
      await kitnetRegistry.registrarKitnet("uri1");
      await kitnetRegistry.registrarKitnet("uri2");
      
      expect(await kitnetRegistry.ownerOf(1)).to.equal(dono.address);
      expect(await kitnetRegistry.ownerOf(2)).to.equal(dono.address);
    });
  });

  describe("Locação (Aluguel)", function () {
    const kitnetId = 1;
    const duracao = 3600; // 1 hora

    beforeEach(async function () {
      await kitnetRegistry.registrarKitnet("ipfs://QmTest123");
    });

    it("Deve iniciar um aluguel com sucesso", async function () {
      const tx = await kitnetRegistry.criarAluguel(kitnetId, locatario.address, duracao);
      const receipt = await tx.wait();
      const bloco = await hre.ethers.provider.getBlock(receipt.blockNumber);
      const tempoFimEsperado = bloco.timestamp + duracao;

      await expect(tx)
        .to.emit(kitnetRegistry, "AluguelIniciado")
        .withArgs(kitnetId, locatario.address, tempoFimEsperado);

      expect(await kitnetRegistry.estaAlugada(kitnetId)).to.equal(true);
      
      const locacao = await kitnetRegistry.alugueisAtivos(kitnetId);
      expect(locacao.locatario).to.equal(locatario.address);
      expect(locacao.estaAtivo).to.equal(true);
    });

    it("Não deve permitir alugar uma kitnet que já está alugada", async function () {
      await kitnetRegistry.criarAluguel(kitnetId, locatario.address, duracao);
      await expect(
        kitnetRegistry.criarAluguel(kitnetId, outraConta.address, duracao)
      ).to.be.revertedWith("Kitnet ja esta alugada");
    });

    it("Apenas o dono da kitnet deve poder criar uma locação", async function () {
      await expect(
        kitnetRegistry.connect(outraConta).criarAluguel(kitnetId, locatario.address, duracao)
      ).to.be.revertedWith("Nao e o dono da kitnet");
    });

    it("Não deve permitir locatário com endereço zero", async function () {
      await expect(
        kitnetRegistry.criarAluguel(kitnetId, hre.ethers.ZeroAddress, duracao)
      ).to.be.revertedWith("Endereco de locatario invalido");
    });
  });

  describe("Transferência de Propriedade", function () {
    const kitnetId = 1;

    beforeEach(async function () {
      await kitnetRegistry.registrarKitnet("ipfs://QmTest123");
    });

    it("Deve impedir a transferência de uma kitnet alugada", async function () {
      await kitnetRegistry.criarAluguel(kitnetId, locatario.address, 3600);
      
      await expect(
        kitnetRegistry.transferFrom(dono.address, outraConta.address, kitnetId)
      ).to.be.revertedWith("Nao e possivel transferir a propriedade com a kitnet alugada");
    });

    it("Deve permitir a transferência de uma kitnet que não está alugada", async function () {
      await kitnetRegistry.transferFrom(dono.address, outraConta.address, kitnetId);
      expect(await kitnetRegistry.ownerOf(kitnetId)).to.equal(outraConta.address);
    });
  });

  describe("Finalização de Aluguel", function () {
    const kitnetId = 1;

    beforeEach(async function () {
      await kitnetRegistry.registrarKitnet("ipfs://QmTest123");
      await kitnetRegistry.criarAluguel(kitnetId, locatario.address, 3600);
    });

    it("Deve permitir que o dono encerre o aluguel", async function () {
      await expect(kitnetRegistry.encerrarAluguel(kitnetId))
        .to.emit(kitnetRegistry, "AluguelEncerrado")
        .withArgs(kitnetId, locatario.address);
        
      expect(await kitnetRegistry.estaAlugada(kitnetId)).to.equal(false);
    });

    it("Deve permitir que o locatário encerre o aluguel", async function () {
      await expect(kitnetRegistry.connect(locatario).encerrarAluguel(kitnetId))
        .to.emit(kitnetRegistry, "AluguelEncerrado")
        .withArgs(kitnetId, locatario.address);
        
      expect(await kitnetRegistry.estaAlugada(kitnetId)).to.equal(false);
    });

    it("Não deve permitir que terceiros encerrem o aluguel", async function () {
      await expect(
        kitnetRegistry.connect(outraConta).encerrarAluguel(kitnetId)
      ).to.be.revertedWith("Apenas o dono ou o locatario podem encerrar o aluguel");
    });

    it("Não deve permitir encerrar um aluguel que já foi finalizado", async function () {
      await kitnetRegistry.encerrarAluguel(kitnetId);
      await expect(
        kitnetRegistry.encerrarAluguel(kitnetId)
      ).to.be.revertedWith("Nao ha aluguel ativo para esta kitnet");
    });
  });

  describe("Consultas (Status)", function () {
    const kitnetId = 1;
    const metadataURI = "ipfs://QmTest123";

    beforeEach(async function () {
      await kitnetRegistry.registrarKitnet(metadataURI);
    });

    it("Deve retornar o status correto da kitnet", async function () {
      let status = await kitnetRegistry.obterStatusKitnet(kitnetId);
      expect(status.proprietario).to.equal(dono.address);
      expect(status.metadataURI).to.equal(metadataURI);
      expect(status.alugadaAtualmente).to.equal(false);

      await kitnetRegistry.criarAluguel(kitnetId, locatario.address, 3600);
      
      status = await kitnetRegistry.obterStatusKitnet(kitnetId);
      expect(status.alugadaAtualmente).to.equal(true);
    });

    it("Deve falhar ao consultar uma kitnet inexistente", async function () {
      await expect(
        kitnetRegistry.obterStatusKitnet(99)
      ).to.be.reverted;
    });
  });
});
