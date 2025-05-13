import React from 'react';
import Modal from '../common/Modal'; // Assuming you have a generic Modal component

import type { Agent, AgentFormData } from '../../types/agent';
import AgentForm from './AgentForm';

interface AgentFormModalProps {
  isOpen: boolean;
  onClose: () => void;
  agentToEdit?: Agent | null;
  onSubmit: (data: AgentFormData) => void;
  isSubmitting?: boolean;
}

const AgentFormModal: React.FC<AgentFormModalProps> = ({
  isOpen,
  onClose,
  agentToEdit,
  onSubmit,
  isSubmitting,
}) => {
  if (!isOpen) return null;

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={agentToEdit ? 'Edit Agent' : 'Add New Agent'}>
      <div className="mt-4">
        <AgentForm
          agent={agentToEdit}
          onSubmit={onSubmit}
          onCancel={onClose}
          isSubmitting={isSubmitting}
        />
      </div>
    </Modal>
  );
};

export default AgentFormModal;