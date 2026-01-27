#!/usr/bin/env bash
# test-validate-network-types.sh
#
# Test suite for validate-network-types.sh script
# Ensures the linter correctly catches concrete network type violations

# Don't exit on error - we expect some tests to fail
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VALIDATOR="${SCRIPT_DIR}/validate-network-types.sh"
TEST_DIR=$(mktemp -d)
PASS=0
FAIL=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

cleanup() {
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

echo "🧪 Testing validate-network-types.sh..."
echo ""

# Test 1: Should PASS - Interface-based code
test_interfaces() {
    cat > "$TEST_DIR/good.go" << 'EOF'
package test
import "net"

func GoodExample(addr net.Addr, conn net.Conn, listener net.Listener) {
    // Using interfaces is correct
}
EOF
    
    if bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 1: Interface-based code passes"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 1: Interface-based code should pass"
        ((FAIL++))
    fi
    rm "$TEST_DIR/good.go"
}

# Test 2: Should FAIL - Concrete *net.TCPAddr
test_tcp_addr() {
    cat > "$TEST_DIR/bad_tcpaddr.go" << 'EOF'
package test
import "net"

func BadTCPAddr(addr *net.TCPAddr) {
    // Should be net.Addr
}
EOF
    
    if ! bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 2: *net.TCPAddr violation detected"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 2: *net.TCPAddr violation not detected"
        ((FAIL++))
    fi
    rm "$TEST_DIR/bad_tcpaddr.go"
}

# Test 3: Should FAIL - Concrete *net.UDPAddr
test_udp_addr() {
    cat > "$TEST_DIR/bad_udpaddr.go" << 'EOF'
package test
import "net"

func BadUDPAddr(addr *net.UDPAddr) {
    // Should be net.Addr
}
EOF
    
    if ! bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 3: *net.UDPAddr violation detected"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 3: *net.UDPAddr violation not detected"
        ((FAIL++))
    fi
    rm "$TEST_DIR/bad_udpaddr.go"
}

# Test 4: Should FAIL - Concrete *net.TCPConn
test_tcp_conn() {
    cat > "$TEST_DIR/bad_tcpconn.go" << 'EOF'
package test
import "net"

func BadTCPConn(conn *net.TCPConn) error {
    return nil
}
EOF
    
    if ! bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 4: *net.TCPConn violation detected"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 4: *net.TCPConn violation not detected"
        ((FAIL++))
    fi
    rm "$TEST_DIR/bad_tcpconn.go"
}

# Test 5: Should FAIL - Concrete *net.UDPConn
test_udp_conn() {
    cat > "$TEST_DIR/bad_udpconn.go" << 'EOF'
package test
import "net"

func BadUDPConn(conn *net.UDPConn) error {
    return nil
}
EOF
    
    if ! bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 5: *net.UDPConn violation detected"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 5: *net.UDPConn violation not detected"
        ((FAIL++))
    fi
    rm "$TEST_DIR/bad_udpconn.go"
}

# Test 6: Should FAIL - Concrete *net.TCPListener
test_tcp_listener() {
    cat > "$TEST_DIR/bad_listener.go" << 'EOF'
package test
import "net"

func BadListener(l *net.TCPListener) error {
    return nil
}
EOF
    
    if ! bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 6: *net.TCPListener violation detected"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 6: *net.TCPListener violation not detected"
        ((FAIL++))
    fi
    rm "$TEST_DIR/bad_listener.go"
}

# Test 7: Should PASS - Comments with concrete types
test_comments() {
    cat > "$TEST_DIR/comments.go" << 'EOF'
package test
import "net"

// This comment mentions *net.TCPAddr but is just documentation
func GoodWithComment(addr net.Addr) {
    // Uses net.Addr interface instead of *net.TCPConn
}
EOF
    
    if bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 7: Comments with concrete types ignored"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 7: Comments should not trigger violations"
        ((FAIL++))
    fi
    rm "$TEST_DIR/comments.go"
}

# Test 8: Should PASS - Test mocks
test_mocks() {
    cat > "$TEST_DIR/mock_test.go" << 'EOF'
package test

type mockAddr struct {
    network string
    address string
}

func (m *mockAddr) Network() string { return m.network }
func (m *mockAddr) String() string  { return m.address }
EOF
    
    if bash "$VALIDATOR" "$TEST_DIR" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Test 8: Test mocks (mockAddr) allowed"
        ((PASS++))
    else
        echo -e "${RED}✗${NC} Test 8: Test mocks should be allowed"
        ((FAIL++))
    fi
    rm "$TEST_DIR/mock_test.go"
}

# Run all tests
test_interfaces
test_tcp_addr
test_udp_addr
test_tcp_conn
test_udp_conn
test_tcp_listener
test_comments
test_mocks

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Test Results: ${GREEN}${PASS} passed${NC}, ${RED}${FAIL} failed${NC}"

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All tests passed!"
    exit 0
else
    echo -e "${RED}✗${NC} Some tests failed"
    exit 1
fi
