// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";

/**
 * @title KitnetRegistryERC721
 * @dev Smart contract for registering and managing kitnets and rental contracts using ERC-721 standard.
 * Based on the TCC proposal "Blockchain para cadastro de kitnet".
 */
contract KitnetRegistry is ERC721URIStorage {
    uint256 private _nextTokenId;

    struct Lease {
        address tenant;
        uint256 startTime;
        uint256 endTime;
        bool isActive;
    }

    // Mapping from kitnet ID to its active lease
    mapping(uint256 => Lease) public activeLeases;
    // Mapping from kitnet ID to its rented status
    mapping(uint256 => bool) public isRented;

    // Events for transparency and auditing
    event KitnetRegistered(uint256 indexed kitnetId, address indexed owner, string metadataURI);
    event LeaseStarted(uint256 indexed kitnetId, address indexed tenant, uint256 endTime);
    event LeaseTerminated(uint256 indexed kitnetId, address indexed tenant);

    constructor() ERC721("KitnetRegistry", "KTNT") {}

    modifier onlyKitnetOwner(uint256 _kitnetId) {
        require(ownerOf(_kitnetId) == msg.sender, "Not the kitnet owner");
        _;
    }

    /**
     * @dev Registers a new kitnet in the system as an NFT.
     * @param _metadataURI The IPFS hash containing the kitnet's digital representation.
     */
    function registerKitnet(string memory _metadataURI) public returns (uint256) {
        _nextTokenId++;
        uint256 newKitnetId = _nextTokenId;

        _mint(msg.sender, newKitnetId);
        _setTokenURI(newKitnetId, _metadataURI);

        emit KitnetRegistered(newKitnetId, msg.sender, _metadataURI);
        return newKitnetId;
    }

    /**
     * @dev Overrides standard transfer behavior to prevent transferring while a kitnet is rented.
     * Note: Uses OpenZeppelin v5 `_update` hook.
     */
    function _update(address to, uint256 tokenId, address auth) internal virtual override returns (address) {
        address from = _ownerOf(tokenId);
        // Only prevent transfer if it's not a mint or burn
        if (from != address(0) && to != address(0)) {
            require(!isRented[tokenId], "Cannot transfer ownership while kitnet is rented");
        }
        return super._update(to, tokenId, auth);
    }

    /**
     * @dev Registers a rental contract (lease) for a kitnet.
     * @param _kitnetId The ID of the kitnet to be rented.
     * @param _tenant The address of the tenant.
     * @param _duration Duration of the lease in seconds.
     */
    function createLease(uint256 _kitnetId, address _tenant, uint256 _duration) public onlyKitnetOwner(_kitnetId) {
        require(!isRented[_kitnetId], "Kitnet already rented");
        require(_tenant != address(0), "Invalid tenant address");

        uint256 endTime = block.timestamp + _duration;
        
        activeLeases[_kitnetId] = Lease({
            tenant: _tenant,
            startTime: block.timestamp,
            endTime: endTime,
            isActive: true
        });

        isRented[_kitnetId] = true;

        emit LeaseStarted(_kitnetId, _tenant, endTime);
    }

    /**
     * @dev Terminates an active rental contract.
     * @param _kitnetId The ID of the kitnet.
     */
    function terminateLease(uint256 _kitnetId) public {
        address owner = ownerOf(_kitnetId); // implicitly checks if token exists
        Lease storage lease = activeLeases[_kitnetId];
        
        require(lease.isActive, "No active lease for this kitnet");
        require(
            msg.sender == owner || msg.sender == lease.tenant,
            "Only owner or tenant can terminate the lease"
        );

        address tenant = lease.tenant;
        lease.isActive = false;
        isRented[_kitnetId] = false;

        emit LeaseTerminated(_kitnetId, tenant);
    }

    /**
     * @dev Retrieves kitnet information. The owner is built into ERC721.
     * @param _kitnetId The ID of the kitnet.
     */
    function getKitnetStatus(uint256 _kitnetId) public view returns (
        address owner,
        string memory metadataURI,
        bool currentlyRented
    ) {
        // ownerOf implicitly checks if token exists
        return (ownerOf(_kitnetId), tokenURI(_kitnetId), isRented[_kitnetId]);
    }
}
