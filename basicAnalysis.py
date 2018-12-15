# goal: find number of unique domains, unqiue effective second level domains, percentage of different answer types
import sys
import ujson
# use https://leetcode.com/problems/implement-trie-prefix-tree/discuss/58989/My-python-solution
class TrieNode:
    # Initialize your data structure here.
    def __init__(self):
        self.word=False
        self.children={}

class Trie:

    def __init__(self):
        self.root = TrieNode()

    # @param {string} word
    # @return {void}
    # Inserts a word into the trie.
    def insert(self, word):
        node=self.root
        for i in word:
            if i not in node.children:
                node.children[i]=TrieNode()
            node=node.children[i]
        exist = node.word
        node.word=True
        return not exist

    # @param {string} word
    # @return {boolean}
    # Returns if the word is in the trie.
    def search(self, word):
        node=self.root
        for i in word:
            if i not in node.children:
                return False
            node=node.children[i]
        return node.word

    # @param {string} prefix
    # @return {boolean}
    # Returns if there is any word in the trie
    # that starts with the given prefix.
    def startsWith(self, prefix):
        node=self.root
        for i in prefix:
            if i not in node.children:
                return False
            node=node.children[i]
        return True


# Your Trie object will be instantiated and called as such:
# trie = Trie()
# trie.insert('somestring')
# trie.search('key')

# choose ujson to load json strings, find https://artem.krylysov.com/blog/2015/09/29/benchmark-python-json-libraries/

if len(sys.argv) != 2:
    print ('Wrong command format.')
    pass
totalDomainCount = 0
unqiueDomainCount = 0
uniqueSecondLevelDomainCount = 0
emptyAnswersCount = 0

#answerTypes = ['A', 'AAAA', 'ANY', 'AXFR', 'CAA', 'CNAME', 'DMARC', 'MX', 'NS', 'PTR', 'TXT', 'SOA', 'SPF']
answersDict = {}
#for answerType in answerTypes:
#    answersDict[answerType] = 0
reverseDomain = Trie()
reverseSecondLevelDomain = Trie()
with open(sys.argv[1]) as f:
    for line in f:
        totalDomainCount += 1
        try:
            jsonObj = ujson.loads(line)
            domainName = jsonObj['name']
            if (reverseDomain.insert(reversed(domainName))):
                unqiueDomainCount += 1
            domainNameParts = domainName.split('.')
            if len(domainNameParts) < 2:
                continue
            if (reverseSecondLevelDomain.insert(reversed(domainNameParts[-2]+'.'+domainNameParts[-1]))):
                uniqueSecondLevelDomainCount += 1
            if jsonObj['data']['answers'] == None or len(jsonObj['data']['answers']) == 0:
                emptyAnswersCount += 1
            else :
                for answer in jsonObj['data']['answers']:
                    type = ''
                    #print (answer)
                    if 'type' in answer:
                        type = answer['type']
                    elif 'Answer' in answer and 'type' in answer['Answer']:
                        type = answer['Answer']['type']
                    else :
                        continue
                    if type not in answersDict.keys():
                        answersDict[type] = 0
                    answersDict[type] += 1
        except Exception as e:
            print ('error dealing with ' + line)
            print ('incur error: ' + str(e))

print ('totalDomainCount: ' + str(totalDomainCount))
print ('unqiueDomainCount: ' + str(unqiueDomainCount))
print ('uniqueSecondLevelDomainCount: ' + str(uniqueSecondLevelDomainCount))
p = emptyAnswersCount * 1.0 / totalDomainCount * 100.0
print ('emptyAnswersCount: ' + str(p)[:6] + '%')
for k,v in answersDict.items():
    p = v * 1.0 / totalDomainCount * 100.0
    print (k + ': ' + str(p)[:6] + '%')
