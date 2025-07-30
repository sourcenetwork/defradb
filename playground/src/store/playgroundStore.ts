import { create } from 'zustand';
import { subscribeWithSelector } from 'zustand/middleware';

export interface PolicyResult {
  message: string;
  type: 'success' | 'error' | 'info';
}

export interface SchemaResult {
  message: string;
  type: 'success' | 'error' | 'info';
}

export interface RelationshipResult {
  message: string;
  type: 'success' | 'error' | 'info';
}

export interface KeypairResult {
  message: string;
  type: 'success' | 'error' | 'info';
}

export interface ClientResult {
  message: string;
  type: 'success' | 'error' | 'info';
}

interface RelationshipData {
  collectionName: string;
  docID: string;
  relation: string;
  targetActor: string;
}

type ClientStatus = 'initializing' | 'ready' | 'error' | 'not_available';

interface AppState {
  client: {
    status: ClientStatus;
    isInitialized: boolean;
    isInitializing: boolean;
    isSourceHubAvailable: boolean;
    result: ClientResult | null;
  };

  policyId: string;

  policy: {
    text: string;
    isLoading: boolean;
    result: PolicyResult | null;
  };

  schema: {
    text: string;
    isLoading: boolean;
    result: SchemaResult | null;
  };

  relationship: {
    data: RelationshipData;
    isLoading: {
      add: boolean;
      delete: boolean;
      verify: boolean;
    };
    result: RelationshipResult | null;
  };

  keypair: {
    isLoading: boolean;
    result: KeypairResult | null;
  };

  setPolicyId: (id: string) => void;
  setPolicyText: (text: string) => void;
  setPolicyLoading: (loading: boolean) => void;
  setPolicyResult: (result: PolicyResult | null) => void;
  addPolicy: () => Promise<void>;

  setSchemaText: (text: string) => void;
  setSchemaLoading: (loading: boolean) => void;
  setSchemaResult: (result: SchemaResult | null) => void;
  addSchema: () => Promise<void>;

  setRelationshipData: (data: Partial<RelationshipData>) => void;
  setRelationshipLoading: (type: 'add' | 'delete' | 'verify', loading: boolean) => void;
  setRelationshipResult: (result: RelationshipResult | null) => void;
  addRelationship: () => Promise<void>;
  deleteRelationship: () => Promise<void>;
  verifyAccess: () => Promise<void>;

  setKeypairLoading: (loading: boolean) => void;
  setKeypairResult: (result: KeypairResult | null) => void;
  resetKeypair: () => Promise<void>;

  setClientStatus: (status: ClientStatus) => void;
  setClientInitialized: (initialized: boolean) => void;
  setClientInitializing: (initializing: boolean) => void;
  setSourceHubAvailable: (available: boolean) => void;
  setClientResult: (result: ClientResult | null) => void;
  initializeClient: () => Promise<void>;
  checkSourceHubAvailability: () => Promise<boolean>;
}

const DEFAULT_POLICY = `name: Test Policy
description: A test policy for playground

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + collaborator
      update:
        expr: owner + collaborator
      delete:
        expr: owner + collaborator

    relations:
      owner:
        types:
          - actor
      collaborator:
        types:
          - actor`;

const DEFAULT_SCHEMA = `type Users @policy(
  id: "policy_id",
  resource: "users"
) {
  name: String
  age: Int
}`;

const DEFAULT_RELATIONSHIP: RelationshipData = {
  collectionName: 'Users',
  docID: 'document_id',
  relation: 'collaborator',
  targetActor: 'did:key:alice',
};

export const usePlaygroundStore = create<AppState>()(
  subscribeWithSelector((set, get) => ({
    client: {
      status: 'not_available' as ClientStatus,
      isInitialized: false,
      isInitializing: false,
      isSourceHubAvailable: false,
      result: null,
    },

    policyId: 'policy_id',

    policy: {
      text: DEFAULT_POLICY,
      isLoading: false,
      result: null,
    },

    schema: {
      text: DEFAULT_SCHEMA,
      isLoading: false,
      result: null,
    },

    relationship: {
      data: DEFAULT_RELATIONSHIP,
      isLoading: { add: false, delete: false, verify: false },
      result: null,
    },

    keypair: {
      isLoading: false,
      result: null,
    },

    setPolicyId: (id: string) => {
      set({ policyId: id });
      const { schema } = get();
      const updatedSchemaText = schema.text.replace(/id: "policy_id"/g, `id: "${id}"`);
      set({ schema: { ...schema, text: updatedSchemaText } });
    },

    setPolicyText: (text: string) =>
      set((state) => ({
        policy: { ...state.policy, text },
      })),

    setPolicyLoading: (loading: boolean) =>
      set((state) => ({
        policy: { ...state.policy, isLoading: loading },
      })),

    setPolicyResult: (result: PolicyResult | null) =>
      set((state) => ({
        policy: { ...state.policy, result },
      })),

    addPolicy: async () => {
      const { policy } = get();
      let client = null;
      if (typeof window !== 'undefined' && window.defradbClient) {
        client = window.defradbClient;
      }
      if (!client) {
        const errorResult: PolicyResult = {
          message: 'Error: Client not initialized',
          type: 'error',
        };
        set((state) => ({
          policy: { ...state.policy, result: errorResult },
        }));
        return;
      }
      set((state) => ({
        policy: {
          ...state.policy,
          isLoading: true,
          result: { message: 'Adding policy...', type: 'info' },
        },
      }));
      try {
        const nodeIdentity = await client.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };
        const response = await client.addDACPolicy(policy.text, context);
        const successMessage = `Policy created successfully: ${JSON.stringify(response, null, 2)}`;
        const successResult: PolicyResult = {
          message: successMessage,
          type: 'success',
        };
        set((state) => ({
          policy: {
            ...state.policy,
            result: successResult,
            isLoading: false,
          },
        }));
        if (response?.PolicyID) {
          get().setPolicyId(response.PolicyID);
        }
      } catch (error) {
        const errorMessage = `Error adding policy: ${error instanceof Error ? error.message : String(error)}`;
        const errorResult: PolicyResult = {
          message: errorMessage,
          type: 'error',
        };
        set((state) => ({
          policy: {
            ...state.policy,
            result: errorResult,
            isLoading: false,
          },
        }));
      }
    },

    setSchemaText: (text: string) =>
      set((state) => ({
        schema: { ...state.schema, text },
      })),

    setSchemaLoading: (loading: boolean) =>
      set((state) => ({
        schema: { ...state.schema, isLoading: loading },
      })),

    setSchemaResult: (result: SchemaResult | null) =>
      set((state) => ({
        schema: { ...state.schema, result },
      })),

    addSchema: async () => {
      const { schema } = get();
      let client = null;
      if (typeof window !== 'undefined' && window.defradbClient) {
        client = window.defradbClient;
      }
      if (!client) {
        const errorResult: SchemaResult = {
          message: 'Error: Client not initialized',
          type: 'error',
        };
        set((state) => ({
          schema: { ...state.schema, result: errorResult },
        }));
        return;
      }
      set((state) => ({
        schema: {
          ...state.schema,
          isLoading: true,
          result: { message: 'Adding schema...', type: 'info' },
        },
      }));
      try {
        const response = await client.addSchema(schema.text);
        const successMessage = `Schema added successfully: ${JSON.stringify(response, null, 2)}`;
        const successResult: SchemaResult = {
          message: successMessage,
          type: 'success',
        };
        set((state) => ({
          schema: {
            ...state.schema,
            result: successResult,
            isLoading: false,
          },
        }));
      } catch (error) {
        const errorMessage = `Error adding schema: ${error instanceof Error ? error.message : String(error)}`;
        const errorResult: SchemaResult = {
          message: errorMessage,
          type: 'error',
        };
        set((state) => ({
          schema: {
            ...state.schema,
            result: errorResult,
            isLoading: false,
          },
        }));
      }
    },

    setRelationshipData: (data: Partial<RelationshipData>) =>
      set((state) => ({
        relationship: {
          ...state.relationship,
          data: { ...state.relationship.data, ...data },
        },
      })),

    setRelationshipLoading: (type: 'add' | 'delete' | 'verify', loading: boolean) =>
      set((state) => ({
        relationship: {
          ...state.relationship,
          isLoading: { ...state.relationship.isLoading, [type]: loading },
        },
      })),

    setRelationshipResult: (result: RelationshipResult | null) =>
      set((state) => ({
        relationship: { ...state.relationship, result },
      })),

    addRelationship: async () => {
      const { relationship } = get();
      let client = null;
      if (typeof window !== 'undefined' && window.defradbClient) {
        client = window.defradbClient;
      }
      if (!client) {
        const errorResult: RelationshipResult = {
          message: 'Error: Client not initialized',
          type: 'error',
        };
        set((state) => ({
          relationship: { ...state.relationship, result: errorResult },
        }));
        return;
      }
      set((state) => ({
        relationship: {
          ...state.relationship,
          isLoading: { add: true, delete: false, verify: false },
          result: { message: 'Adding relationship...', type: 'info' },
        },
      }));
      try {
        const nodeIdentity = await client.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };
        const response = await client.addDACActorRelationship(
          relationship.data.collectionName,
          relationship.data.docID,
          relationship.data.relation,
          relationship.data.targetActor,
          context,
        );
        const successMessage = `Relationship added successfully: ${JSON.stringify(response, null, 2)}`;
        const successResult: RelationshipResult = {
          message: successMessage,
          type: 'success',
        };
        set((state) => ({
          relationship: {
            ...state.relationship,
            result: successResult,
            isLoading: { add: false, delete: false, verify: false },
          },
        }));
      } catch (error) {
        const errorMessage = `Error adding relationship: ${error instanceof Error ? error.message : String(error)}`;
        const errorResult: RelationshipResult = {
          message: errorMessage,
          type: 'error',
        };
        set((state) => ({
          relationship: {
            ...state.relationship,
            result: errorResult,
            isLoading: { add: false, delete: false, verify: false },
          },
        }));
      }
    },

    deleteRelationship: async () => {
      const { relationship } = get();
      let client = null;
      if (typeof window !== 'undefined' && window.defradbClient) {
        client = window.defradbClient;
      }
      if (!client) {
        const errorResult: RelationshipResult = {
          message: 'Error: Client not initialized',
          type: 'error',
        };
        set((state) => ({
          relationship: { ...state.relationship, result: errorResult },
        }));
        return;
      }
      set((state) => ({
        relationship: {
          ...state.relationship,
          isLoading: { add: false, delete: true, verify: false },
          result: { message: 'Deleting relationship...', type: 'info' },
        },
      }));
      try {
        const nodeIdentity = await client.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };
        const response = await client.deleteDACActorRelationship(
          relationship.data.collectionName,
          relationship.data.docID,
          relationship.data.relation,
          relationship.data.targetActor,
          context,
        );
        const successMessage = `Relationship deleted successfully: ${JSON.stringify(response, null, 2)}`;
        const successResult: RelationshipResult = {
          message: successMessage,
          type: 'success',
        };
        set((state) => ({
          relationship: {
            ...state.relationship,
            result: successResult,
            isLoading: { add: false, delete: false, verify: false },
          },
        }));
      } catch (error) {
        const errorMessage = `Error deleting relationship: ${error instanceof Error ? error.message : String(error)}`;
        const errorResult: RelationshipResult = {
          message: errorMessage,
          type: 'error',
        };
        set((state) => ({
          relationship: {
            ...state.relationship,
            result: errorResult,
            isLoading: { add: false, delete: false, verify: false },
          },
        }));
      }
    },

    verifyAccess: async () => {
      const { relationship, policyId } = get();
      let client = null;
      if (typeof window !== 'undefined' && window.defradbClient) {
        client = window.defradbClient;
      }
      if (!client) {
        const errorResult: RelationshipResult = {
          message: 'Error: Client not initialized',
          type: 'error',
        };
        set((state) => ({
          relationship: { ...state.relationship, result: errorResult },
        }));
        return;
      }
      if (!policyId || policyId === 'policy_id') {
        const errorMessage = 'Error: Policy ID not available. Please add a policy first.';
        const errorResult: RelationshipResult = {
          message: errorMessage,
          type: 'error',
        };
        set((state) => ({
          relationship: { ...state.relationship, result: errorResult },
        }));
        return;
      }
      set((state) => ({
        relationship: {
          ...state.relationship,
          isLoading: { add: false, delete: false, verify: true },
          result: { message: 'Verifying access...', type: 'info' },
        },
      }));
      try {
        const nodeIdentity = await client.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };
        const response = await client.verifyDACAccess(
          'read',
          relationship.data.targetActor,
          policyId,
          relationship.data.collectionName.toLowerCase(),
          relationship.data.docID,
          context,
        );
        const successMessage = `Access verification result: ${JSON.stringify(response, null, 2)}`;
        const successResult: RelationshipResult = {
          message: successMessage,
          type: 'success',
        };
        set((state) => ({
          relationship: {
            ...state.relationship,
            result: successResult,
            isLoading: { add: false, delete: false, verify: false },
          },
        }));
      } catch (error) {
        const errorMessage = `Error verifying access: ${error instanceof Error ? error.message : String(error)}`;
        const errorResult: RelationshipResult = {
          message: errorMessage,
          type: 'error',
        };
        set((state) => ({
          relationship: {
            ...state.relationship,
            result: errorResult,
            isLoading: { add: false, delete: false, verify: false },
          },
        }));
      }
    },

    setKeypairLoading: (loading: boolean) =>
      set((state) => ({
        keypair: { ...state.keypair, isLoading: loading },
      })),

    setKeypairResult: (result: KeypairResult | null) =>
      set((state) => ({
        keypair: { ...state.keypair, result },
      })),

    resetKeypair: async () => {
      set((state) => ({
        keypair: {
          ...state.keypair,
          isLoading: true,
          result: { message: 'Deleting keypair...', type: 'info' },
        },
      }));
      try {
        const acpDeleteKeypair = (window as any).acp_DeleteKeypair;
        if (acpDeleteKeypair) {
          const error = await acpDeleteKeypair();
          if (error) {
            const errorResult: KeypairResult = {
              message: `Error deleting keypair: ${error.message ?? error}`,
              type: 'error',
            };
            set((state) => ({
              keypair: {
                ...state.keypair,
                result: errorResult,
                isLoading: false,
              },
            }));
          } else {
            const successResult: KeypairResult = {
              message: 'Keypair reset successfully! Reloading the page...',
              type: 'success',
            };
            set((state) => ({
              keypair: {
                ...state.keypair,
                result: successResult,
                isLoading: false,
              },
            }));
            setTimeout(() => {
              window.location.reload();
            }, 2000);
          }
        } else {
          const errorResult: KeypairResult = {
            message: 'Error: acp_DeleteKeypair function not found',
            type: 'error',
          };
          set((state) => ({
            keypair: {
              ...state.keypair,
              result: errorResult,
              isLoading: false,
            },
          }));
        }
      } catch (error) {
        const errorResult: KeypairResult = {
          message: `Error deleting keypair: ${error instanceof Error ? error.message : String(error)}`,
          type: 'error',
        };
        set((state) => ({
          keypair: {
            ...state.keypair,
            result: errorResult,
            isLoading: false,
          },
        }));
      }
    },

    setClientStatus: (status: ClientStatus) =>
      set((state) => ({
        client: { ...state.client, status },
      })),

    setClientInitialized: (initialized: boolean) =>
      set((state) => ({
        client: { ...state.client, isInitialized: initialized },
      })),

    setClientInitializing: (initializing: boolean) =>
      set((state) => ({
        client: { ...state.client, isInitializing: initializing },
      })),

    setSourceHubAvailable: (available: boolean) =>
      set((state) => ({
        client: { ...state.client, isSourceHubAvailable: available },
      })),

    setClientResult: (result: ClientResult | null) =>
      set((state) => ({
        client: { ...state.client, result },
      })),

    checkSourceHubAvailability: async (): Promise<boolean> => {
      try {
        const response = await fetch('/api/cosmos/base/tendermint/v1beta1/node_info');
        const isAvailable = response.ok;
        set((state) => ({
          client: { ...state.client, isSourceHubAvailable: isAvailable },
        }));
        return isAvailable;
      } catch (error) {
        console.log('SourceHub not available:', error);
        set((state) => ({
          client: { ...state.client, isSourceHubAvailable: false },
        }));
        return false;
      }
    },

    initializeClient: async () => {
      const { client } = get();
      if (client.isInitialized || client.isInitializing) {
        return;
      }
      set((state) => ({
        client: {
          ...state.client,
          isInitializing: true,
          status: 'initializing',
          result: { message: 'Initializing DefraDB client...', type: 'info' },
        },
      }));
      try {
        if (!window.defradb) {
          set((state) => ({
            client: {
              ...state.client,
              status: 'error',
              isInitializing: false,
              result: {
                message: 'DefraDB WASM module not available',
                type: 'error',
              },
            },
          }));
          return;
        }
        const acpClient = import.meta.env.VITE_ACP_CLIENT;
        let useSourceHub = false;
        if (acpClient === 'sourcehub') {
          useSourceHub = await get().checkSourceHubAvailability();
        }
        const db = useSourceHub
          ? await window.defradb.open('sourcehub')
          : await window.defradb.open();
        window.defradbClient = db;
        set((state) => ({
          client: {
            ...state.client,
            status: 'ready',
            isInitialized: true,
            isInitializing: false,
            isSourceHubAvailable: useSourceHub,
            result: {
              message: `DefraDB client initialized with ${useSourceHub ? 'SourceHub ACP' : 'Local ACP'}`,
              type: 'success',
            },
          },
        }));
        console.log('DefraDB client initialized with', useSourceHub ? 'SourceHub ACP' : 'Local ACP');
      } catch (error) {
        console.error('Failed to initialize DefraDB client:', error);
        set((state) => ({
          client: {
            ...state.client,
            status: 'error',
            isInitializing: false,
            result: {
              message: `Failed to initialize DefraDB client: ${error instanceof Error ? error.message : String(error)}`,
              type: 'error',
            },
          },
        }));
      }
    },
  })),
); 