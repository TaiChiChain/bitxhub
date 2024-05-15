pragma solidity >=0.7.0 <0.9.0;

    struct UnlockingRecord {
        uint256 Amount;
        uint64 UnlockTimestamp;
    }

    struct LiquidStakingTokenInfo {
        uint64 PoolID;
        uint256 Principal;
        uint256 Unlocked;
        uint64 ActiveEpoch;
        UnlockingRecord[] UnlockingRecords;
    }

// compatible with ERC721
interface LiquidStakingToken {
    // ERC721

    event Transfer(address indexed _from, address indexed _to, uint256 indexed _tokenId);
    event Approval(address indexed _owner, address indexed _approved, uint256 indexed _tokenId);
    event ApprovalForAll(address indexed _owner, address indexed _operator, bool _approved);
    event UpdateInfo(uint256 indexed _tokenId, uint256 newPrincipal, uint256 newUnlocked, uint64 newActiveEpoch);

    function balanceOf(address _owner) external view returns (uint256);

    function ownerOf(uint256 _tokenId) external view returns (address);

    function safeTransferFrom(address _from, address _to, uint256 _tokenId, bytes calldata data) external payable;

    function safeTransferFrom(address _from, address _to, uint256 _tokenId) external payable;

    function transferFrom(address _from, address _to, uint256 _tokenId) external payable;

    function approve(address _approved, uint256 _tokenId) external payable;

    function setApprovalForAll(address _operator, bool _approved) external;

    function getApproved(uint256 _tokenId) external view returns (address);

    function isApprovedForAll(address _owner, address _operator) external view returns (bool);

    // liquid staking token specific methods

    function getInfo(uint256 _tokenId) external view returns (LiquidStakingTokenInfo memory info);

    function getLockedReward(uint256 _tokenId) external view returns (uint256);

    function getUnlockingCoin(uint256 _tokenId) external view returns (uint256);

    function getUnlockedCoin(uint256 _tokenId) external view returns (uint256);

    function getLockedCoin(uint256 _tokenId) external view returns (uint256);

    function getTotalCoin(uint256 _tokenId) external view returns (uint256);
}