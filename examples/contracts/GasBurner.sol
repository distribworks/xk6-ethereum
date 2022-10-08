// SPDX-License-Identifier: MIT

pragma solidity ^0.5.0;

contract GasBurner {
     uint256 total;
     function burnGas(uint limit) public {
         for(uint256 i = 0; i < limit; i++) {
             total++;
         }
     }

     function getTotal() public view returns(uint256) {
         return total;
     }
}
