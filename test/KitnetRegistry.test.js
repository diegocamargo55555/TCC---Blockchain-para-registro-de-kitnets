import { expect } from "chai";
import hre from "hardhat";

describe("KitnetRegistry", function () {
  let kitnetRegistry;
  let owner;
  let tenant;
  let otherAccount;

  beforeEach(async function () {
    [owner, tenant, otherAccount] = await hre.ethers.getSigners();
    const KitnetRegistry = await hre.ethers.getContractFactory("KitnetRegistry");
    kitnetRegistry = await KitnetRegistry.deploy();
  });

  describe("Registration", function () {
    it("Should register a new kitnet and emit event", async function () {
      const metadataURI = "ipfs://QmTest123";
      
      await expect(kitnetRegistry.registerKitnet(metadataURI))
        .to.emit(kitnetRegistry, "KitnetRegistered")
        .withArgs(1, owner.address, metadataURI);

      expect(await kitnetRegistry.ownerOf(1)).to.equal(owner.address);
      expect(await kitnetRegistry.tokenURI(1)).to.equal(metadataURI);
    });
  });

  describe("Leasing", function () {
    beforeEach(async function () {
      await kitnetRegistry.registerKitnet("ipfs://QmTest123");
    });

    it("Should create a lease successfully", async function () {
      const duration = 3600; // 1 hour
      
      const tx = await kitnetRegistry.createLease(1, tenant.address, duration);
      const receipt = await tx.wait();
      const block = await hre.ethers.provider.getBlock(receipt.blockNumber);
      const expectedEndTime = block.timestamp + duration;

      await expect(tx)
        .to.emit(kitnetRegistry, "LeaseStarted")
        .withArgs(1, tenant.address, expectedEndTime);

      expect(await kitnetRegistry.isRented(1)).to.equal(true);
      
      const lease = await kitnetRegistry.activeLeases(1);
      expect(lease.tenant).to.equal(tenant.address);
      expect(lease.isActive).to.equal(true);
    });

    it("Should prevent transferring a rented kitnet", async function () {
      await kitnetRegistry.createLease(1, tenant.address, 3600);
      
      // Attempt to transfer
      await expect(
        kitnetRegistry.transferFrom(owner.address, otherAccount.address, 1)
      ).to.be.revertedWith("Cannot transfer ownership while kitnet is rented");
    });

    it("Should allow transferring an unrented kitnet", async function () {
      await kitnetRegistry.transferFrom(owner.address, otherAccount.address, 1);
      expect(await kitnetRegistry.ownerOf(1)).to.equal(otherAccount.address);
    });

    it("Should allow owner to terminate lease", async function () {
      await kitnetRegistry.createLease(1, tenant.address, 3600);
      
      await expect(kitnetRegistry.terminateLease(1))
        .to.emit(kitnetRegistry, "LeaseTerminated")
        .withArgs(1, tenant.address);
        
      expect(await kitnetRegistry.isRented(1)).to.equal(false);
    });

    it("Should allow tenant to terminate lease", async function () {
      await kitnetRegistry.createLease(1, tenant.address, 3600);
      
      await expect(kitnetRegistry.connect(tenant).terminateLease(1))
        .to.emit(kitnetRegistry, "LeaseTerminated")
        .withArgs(1, tenant.address);
        
      expect(await kitnetRegistry.isRented(1)).to.equal(false);
    });

    it("Should not allow other account to terminate lease", async function () {
      await kitnetRegistry.createLease(1, tenant.address, 3600);
      
      await expect(
        kitnetRegistry.connect(otherAccount).terminateLease(1)
      ).to.be.revertedWith("Only owner or tenant can terminate the lease");
    });
  });
});
